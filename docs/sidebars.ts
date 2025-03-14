// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  sidebar: [
    "overview",
    {
      type: "category",
      label: "Concepts",
      items: [
        "concepts/agents",
        "concepts/threads",
        "concepts/tools",
        "concepts/users",
        "concepts/oauth-apps",
        "concepts/tasks",
      ],
    },
    {
      type: "category",
      label: "Tutorials",
      items: [
        "tutorials/github-assistant",
        "tutorials/github-ci-failure-notifier",
        "tutorials/knowledge-assistant",
        "tutorials/slack-alerts-assistant",
      ],
    },
    {
      type: "category",
      label: "Self Hosted",
      items: [
        "installation/Installation",
        {
          type: "category",
          label: "Configuration",
          items: [
            "configuration/general",
            "configuration/auth-providers",
            "configuration/email-webhook",
            "configuration/encryption-providers",
            "configuration/model-providers",
            "configuration/workspace-provider",
          ],
        },
        "enterprise",
        {
          type: "category",
          label: "Tools",
          items: [
            "tools/first-tool",
            "tools/integrating-oauth",
          ],
        },
      ],
    },
  ],
};

export default sidebars;
