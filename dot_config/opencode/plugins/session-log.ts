import type { Plugin } from "@opencode-ai/plugin";

// OpenCode has no "session.idle" hook key: session events arrive through the
// generic `event` hook, and the payload is `{ type, properties }`, not `data`.
// The previous shape silently never ran. The working directory comes from the
// plugin input, since EventSessionIdle carries only a sessionID.
export default (async ({ $, directory }) => {
  return {
    event: async ({ event }) => {
      if (event.type !== "session.idle") return;
      const sid = event.properties.sessionID;
      if (!sid) return;
      const dir = directory ?? ".";
      await $`dot agent hook session opencode ${sid} ${dir}`;
      await $`dot agent hook usage opencode ${sid} ${dir}`;
    },
  };
}) satisfies Plugin;
