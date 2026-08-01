import type { Plugin } from "@opencode-ai/plugin";

export default (async ({ $ }) => {
  return {
    "session.idle": async ({ event }) => {
      const sid = event.data?.id;
      const dir = event.data?.directory;
      if (sid) await $`dot agent hook session opencode ${sid} ${dir ?? "."}`;
    },
  };
}) satisfies Plugin;
