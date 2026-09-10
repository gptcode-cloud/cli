(function () {
  "use strict";

  const modelContext = document.modelContext;
  if (!modelContext || typeof modelContext.registerTool !== "function") {
    return;
  }

  const INDEX_PATH = "/webmcp-index.json";
  const MAX_RESULTS = 10;
  const MAX_PAGE_CHARACTERS = 12000;
  let indexPromise;

  async function fetchResource(path, signal) {
    const response = await fetch(path, {
      credentials: "same-origin",
      headers: { Accept: "text/html, application/json, text/plain" },
      signal: signal
    });

    if (!response.ok) {
      throw new Error(`GPTCode resource returned HTTP ${response.status}.`);
    }

    return response;
  }

  function getIndex(signal) {
    if (!indexPromise) {
      indexPromise = fetchResource(INDEX_PATH, signal).then(function (response) {
        return response.json();
      }).catch(function (error) {
        indexPromise = undefined;
        throw error;
      });
    }

    return indexPromise;
  }

  function allEntries(index) {
    return [index.pages, index.posts, index.skills].flat().filter(Boolean);
  }

  function uniqueEntries(entries) {
    const seen = new Set();

    return entries.filter(function (entry) {
      const key = `${entry.section}:${entry.title.toLowerCase()}`;
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    });
  }

  function normalizePath(value) {
    const candidate = new URL(value, window.location.origin);
    if (candidate.origin !== window.location.origin) {
      throw new Error("Only published pages on gptcode.dev can be read.");
    }

    return candidate.pathname.replace(/\/$/, "") || "/";
  }

  function entryPath(entry) {
    return normalizePath(entry.path);
  }

  function searchScore(entry, terms) {
    const title = (entry.title || "").toLowerCase();
    const summary = (entry.summary || "").toLowerCase();
    let score = 0;
    let matchedTerms = 0;

    terms.forEach(function (term) {
      const titleMatch = title.includes(term);
      const summaryMatch = summary.includes(term);
      if (titleMatch || summaryMatch) matchedTerms += 1;
      if (titleMatch) score += 3;
      if (summaryMatch) score += 1;
    });

    return score + (matchedTerms / terms.length) * 10;
  }

  const tools = [
    {
      name: "get_gptcode_overview",
      title: "Get the GPTCode project overview",
      description: "Return the official GPTCode project thesis and links to its documentation, skills, essays, source code, and creator.",
      inputSchema: {
        type: "object",
        properties: {},
        additionalProperties: false
      },
      annotations: {
        readOnlyHint: true,
        idempotentHint: true,
        openWorldHint: false,
        untrustedContentHint: true
      },
      execute: async function (_input, context) {
        const response = await fetchResource("/llms.txt", context.signal);
        return {
          source: new URL("/llms.txt", window.location.origin).href,
          content: await response.text()
        };
      }
    },
    {
      name: "search_gptcode_content",
      title: "Search GPTCode documentation and essays",
      description: "Search the public GPTCode website across documentation pages, blog essays, and engineering skills. Returns matching titles, summaries, sections, and canonical links.",
      inputSchema: {
        type: "object",
        properties: {
          query: {
            type: "string",
            minLength: 2,
            maxLength: 120,
            description: "Words or phrase to find in GPTCode content."
          },
          section: {
            type: "string",
            enum: ["all", "documentation", "blog", "skills"],
            default: "all",
            description: "Optional content section to search."
          },
          limit: {
            type: "integer",
            minimum: 1,
            maximum: MAX_RESULTS,
            default: 5,
            description: "Maximum number of matches to return."
          }
        },
        required: ["query"],
        additionalProperties: false
      },
      annotations: {
        readOnlyHint: true,
        idempotentHint: true,
        openWorldHint: false,
        untrustedContentHint: true
      },
      execute: async function (input, context) {
        const query = input.query.trim().toLowerCase();
        const terms = query.split(/\s+/).filter(Boolean);
        const section = input.section || "all";
        const limit = Math.min(input.limit || 5, MAX_RESULTS);
        const index = await getIndex(context.signal);

        const matches = uniqueEntries(allEntries(index)).map(function (entry) {
          return { entry: entry, score: searchScore(entry, terms) };
        }).filter(function (match) {
          return match.score > 0 && (section === "all" || match.entry.section === section);
        }).sort(function (left, right) {
          return right.score - left.score || left.entry.title.localeCompare(right.entry.title);
        }).slice(0, limit).map(function (match) {
          return {
            title: match.entry.title,
            section: match.entry.section,
            summary: match.entry.summary,
            published: match.entry.published || null,
            link: new URL(match.entry.path, window.location.origin).href
          };
        });

        return { query: input.query, count: matches.length, results: matches };
      }
    },
    {
      name: "read_gptcode_page",
      title: "Read a GPTCode page",
      description: "Read the main text of a page returned by search_gptcode_content. Only indexed, same-origin GPTCode pages are accepted.",
      inputSchema: {
        type: "object",
        properties: {
          path: {
            type: "string",
            minLength: 1,
            maxLength: 240,
            description: "The canonical path or link returned by search_gptcode_content."
          }
        },
        required: ["path"],
        additionalProperties: false
      },
      annotations: {
        readOnlyHint: true,
        idempotentHint: true,
        openWorldHint: false,
        untrustedContentHint: true
      },
      execute: async function (input, context) {
        const requestedPath = normalizePath(input.path);
        const index = await getIndex(context.signal);
        const entry = allEntries(index).find(function (candidate) {
          return entryPath(candidate) === requestedPath;
        });

        if (!entry) {
          throw new Error("The requested path is not in the published GPTCode content index.");
        }

        const response = await fetchResource(entry.path, context.signal);
        const html = await response.text();
        const page = new DOMParser().parseFromString(html, "text/html");
        const main = page.querySelector("main") || page.body;
        main.querySelectorAll("script, style, nav, footer").forEach(function (node) {
          node.remove();
        });

        const content = (main.textContent || "").replace(/\s+/g, " ").trim();
        return {
          title: entry.title,
          section: entry.section,
          source: new URL(entry.path, window.location.origin).href,
          truncated: content.length > MAX_PAGE_CHARACTERS,
          content: content.slice(0, MAX_PAGE_CHARACTERS)
        };
      }
    }
  ];

  tools.forEach(function (tool) {
    Promise.resolve(document.modelContext.registerTool(tool)).catch(function (error) {
      console.warn(`Could not register WebMCP tool ${tool.name}.`, error);
    });
  });
})();
