const assert = require("node:assert/strict");
const test = require("node:test");

const SCRIPT_PATH = require.resolve("../assets/webmcp.js");

function response(body) {
  return {
    ok: true,
    status: 200,
    json: async function () { return body; },
    text: async function () { return body; }
  };
}

function loadTools(index) {
  const tools = [];

  global.document = {
    modelContext: {
      registerTool: function (tool) {
        tools.push(tool);
      }
    }
  };
  global.window = { location: { origin: "https://gptcode.dev" } };
  global.fetch = async function (path) {
    if (path === "/webmcp-index.json") return response(index);
    if (path === "/llms.txt") return response("# GPTCode\n\nEvidence, not just answers.");
    if (path === "/blog/local-first") return response("<main>Local first evidence</main>");
    throw new Error(`Unexpected fetch: ${path}`);
  };
  global.DOMParser = class {
    parseFromString(html) {
      const main = {
        textContent: html.replace(/<[^>]+>/g, " "),
        querySelectorAll: function () { return []; }
      };
      return { body: main, querySelector: function () { return main; } };
    }
  };

  delete require.cache[SCRIPT_PATH];
  require(SCRIPT_PATH);
  return tools;
}

const index = {
  pages: [{
    title: "Commands",
    section: "documentation",
    path: "/reference/commands",
    summary: "Command reference"
  }],
  posts: [{
    title: "When Should a Coding Agent Go to the Cloud?",
    section: "blog",
    path: "/blog/local-first",
    summary: "Local-first routing calibration",
    published: "2026-08-20T00:00:00-03:00"
  }, {
    title: "Agent Routing vs. Tool Search",
    section: "blog",
    path: "/blog/agent-routing-new",
    summary: "Agent routing comparison",
    published: "2025-12-01T00:00:00-03:00"
  }, {
    title: "Agent Routing vs. Tool Search",
    section: "blog",
    path: "/blog/agent-routing-old",
    summary: "Agent routing comparison",
    published: "2025-11-29T00:00:00-03:00"
  }],
  skills: []
};

test("registers a minimal read-only tool surface", function () {
  const tools = loadTools(index);

  assert.deepEqual(tools.map(function (tool) { return tool.name; }), [
    "get_gptcode_overview",
    "search_gptcode_content",
    "read_gptcode_page"
  ]);
  tools.forEach(function (tool) {
    assert.equal(tool.annotations.readOnlyHint, true);
    assert.equal(tool.annotations.untrustedContentHint, true);
  });
});

test("searches indexed content and returns canonical links", async function () {
  const tools = loadTools(index);
  const search = tools.find(function (tool) {
    return tool.name === "search_gptcode_content";
  });

  const result = await search.execute(
    { query: "local-first routing", section: "blog", limit: 3 },
    { signal: new AbortController().signal }
  );

  assert.equal(result.count, 2);
  assert.equal(result.results[0].title, "When Should a Coding Agent Go to the Cloud?");
  assert.equal(result.results[0].link, "https://gptcode.dev/blog/local-first");
  assert.equal(result.results[1].link, "https://gptcode.dev/blog/agent-routing-new");
});

test("reads only an indexed same-origin page", async function () {
  const tools = loadTools(index);
  const read = tools.find(function (tool) {
    return tool.name === "read_gptcode_page";
  });
  const context = { signal: new AbortController().signal };

  await assert.rejects(
    read.execute({ path: "https://example.com/private" }, context),
    /Only published pages on gptcode\.dev/
  );

  const result = await read.execute({ path: "/blog/local-first" }, context);
  assert.equal(result.source, "https://gptcode.dev/blog/local-first");
  assert.match(result.content, /Local first evidence/);
  assert.equal(result.truncated, false);
});
