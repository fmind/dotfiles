# /// script
# requires-python = ">=3.10"
# dependencies = ["google-antigravity"]
# ///
"""Fan work out to static subagents under a hard budget and an explicit policy.

Run with `GEMINI_API_KEY=... uv run orchestrator.py <workspace>`; every knob below
is the orchestration contract, so change it here rather than in the prompt.
"""

import asyncio
import pathlib
import sys

import pydantic
from google.antigravity import Agent, LocalAgentConfig, hooks, policy, types


class Audit(pydantic.BaseModel):
    """Typed contract for the orchestrator's answer."""

    findings: list[str]
    verdict: str


def record_finding(area: str, detail: str) -> str:
    """Records one audit finding for the given area."""
    # Custom Python tools bypass the policy engine entirely -- it gates built-ins
    # such as run_command and edit_file -- so they must stay side-effect free.
    print(f"[{area}] {detail}", file=sys.stderr)
    return "recorded"


@hooks.pre_turn
async def trace_turn(prompt: types.Content) -> types.HookResult:
    """Logs every turn so a long orchestration stays auditable."""
    print(f"turn: {prompt}", file=sys.stderr)
    return types.HookResult(allow=True)


@hooks.on_tool_error
async def survive_tool_error(error: Exception) -> None:
    """Keeps one failing tool call from aborting the whole run."""
    print(f"tool error: {error}", file=sys.stderr)


READER = types.SubagentConfig(
    name="reader",
    description="Reads source files and reports what a module does.",
    system_instructions="Summarize behavior only. Never propose edits.",
    tools=[],
)

CRITIC = types.SubagentConfig(
    name="critic",
    description="Refutes a claimed finding using the source as evidence.",
    system_instructions="Default to refuting. Confirm only with a file and line.",
    tools=[record_finding],
)


def build(workspace: str) -> LocalAgentConfig:
    """Assembles the orchestrator: one parent, two isolated subagents."""
    root = str(pathlib.Path(workspace).resolve())
    return LocalAgentConfig(
        model="gemini-3.8-flash",
        workspaces=[root],
        tools=[record_finding],  # subagent tools must also be registered here
        subagents=[READER, CRITIC],
        capabilities=types.CapabilitiesConfig(
            enable_subagents=True,
            allowed_subagents=["reader", "critic"],
            max_subagent_depth=1,  # without this, delegation nests without bound
            # read_only() omits START_SUBAGENT: filtering to it alone silently
            # disables delegation and rejects max_subagent_depth at validation.
            enabled_tools=[
                *types.BuiltinTools.read_only(),
                types.BuiltinTools.START_SUBAGENT,
            ],
        ),
        # A budget is the only hard stop: policies gate *which* tools may run,
        # never how many tokens a delegating parent spends fanning out.
        budget_config=types.BudgetConfig(max_model_calls=40, max_total_tokens=400_000),
        # Unattended runs must not use policy.safe_defaults(handler): it routes
        # every write to a human and blocks forever with no console attached.
        policies=[
            *[policy.allow(tool.value) for tool in types.BuiltinTools.read_only()],
            *policy.workspace_only([root]),
            policy.deny(types.BuiltinTools.RUN_COMMAND.value),
        ],
        hooks=[trace_turn, survive_tool_error],
        response_schema=Audit,
    )


async def main(workspace: str) -> None:
    """Runs one orchestration and prints the typed result."""
    async with Agent(build(workspace)) as agent:
        response = await agent.chat(
            "Delegate to reader for each top-level package, then have critic refute"
            " every claim. Report only surviving findings."
        )
        data = await response.structured_output()
        print(data["verdict"])
        for finding in data["findings"]:
            print(f"- {finding}")
        if response.usage_metadata:
            print(f"tokens: {response.usage_metadata.total_token_count}", file=sys.stderr)


if __name__ == "__main__":
    asyncio.run(main(sys.argv[1] if len(sys.argv) > 1 else "."))
