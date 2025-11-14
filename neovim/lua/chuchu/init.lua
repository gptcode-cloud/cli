-- chuchu.nvim – generic Neovim integration for Chuchu
--
-- Features:
-- - Detects project type (Elixir / Ruby / Go / TypeScript) via simple heuristics.
-- - Uses corresponding CLI commands:
--     Elixir     -> `chu feature-elixir`
--     Ruby       -> `chu feature-ruby`   (you implement this in Go)
--     Go         -> `chu feature-go`     (you implement this in Go)
--     TypeScript -> `chu feature-ts`     (already scaffolded)
-- - :ChuchuFeature → opens a floating prompt for the feature description.
-- - Renders ```tests / ```impl fenced blocks from stdout and opens a 3-pane layout:
--     left top: tests
--     left bottom: implementation
--     right: conversation (prompt + raw output)
-- - Feedback commands store snapshots in ~/.chuchu/memories.jsonl:
--     :ChuchuFeedbackGood  (default key: <leader>ck)
--     :ChuchuFeedbackBad   (default key: <leader>cx)

local M = {}

local config = {
  feature_cmds = {
    elixir = { "chu", "feature-elixir" },
    ruby   = { "chu", "feature-ruby" },
    go     = { "chu", "feature-go" },
    ts     = { "chu", "feature-ts" },
  },
  keymaps = {
    code          = "<leader>cd",
    verified      = "<leader>vf",
    failed        = "<leader>fr",
    shell_help    = "<leader>xs",
    toggle_chat   = "<leader>cc",
  },
  memory_file = vim.fn.expand("~/.chuchu/memories.jsonl"),
}

local chat_state = { buf = nil, win = nil }

--- Setup to be called from your plugin manager.
-- Example (lazy.nvim):
--   {
--     dir = "~/workspace/chuchu/neovim",
--     config = function()
--       require("chuchu").setup()
--     end,
--   }
function M.setup(opts)
  config = vim.tbl_deep_extend("force", config, opts or {})

  vim.api.nvim_create_user_command("ChuchuCode", function()
    M.start_code_conversation()
  end, {})

  vim.api.nvim_create_user_command("ChuchuVerified", function()
    M.record_feedback("good")
  end, {})

  vim.api.nvim_create_user_command("ChuchuFailed", function()
    M.record_feedback("bad")
  end, {})

  vim.api.nvim_create_user_command("ChuchuShell", function()
    M.shell_help()
  end, {})

  vim.api.nvim_create_user_command("ChuchuToggleChat", function()
    M.toggle_chat()
  end, {})

  local km = config.keymaps
  if km.code and km.code ~= "" then
    vim.keymap.set("n", km.code, ":ChuchuCode<CR>", {
      silent = true,
      noremap = true,
      desc = "Chuchu: generate code",
    })
  end
  if km.verified and km.verified ~= "" then
    vim.keymap.set("n", km.verified, ":ChuchuVerified<CR>", {
      silent = true,
      noremap = true,
      desc = "Chuchu: verified code",
    })
  end
  if km.failed and km.failed ~= "" then
    vim.keymap.set("n", km.failed, ":ChuchuFailed<CR>", {
      silent = true,
      noremap = true,
      desc = "Chuchu: failed code",
    })
  end
  if km.shell_help and km.shell_help ~= "" then
    vim.keymap.set("n", km.shell_help, ":ChuchuShell<CR>", {
      silent = true,
      noremap = true,
      desc = "Chuchu: shell help",
    })
  end
  if km.toggle_chat and km.toggle_chat ~= "" then
    vim.keymap.set("n", km.toggle_chat, ":ChuchuToggleChat<CR>", {
      silent = true,
      noremap = true,
      desc = "Chuchu: toggle chat",
    })
  end
end

function M.toggle_chat()
  if chat_state.win and vim.api.nvim_win_is_valid(chat_state.win) then
    vim.api.nvim_win_close(chat_state.win, true)
    chat_state.win = nil
  elseif chat_state.buf and vim.api.nvim_buf_is_valid(chat_state.buf) then
    vim.cmd("vsplit")
    chat_state.win = vim.api.nvim_get_current_win()
    vim.api.nvim_win_set_buf(chat_state.win, chat_state.buf)
  end
end

function M.shell_help()
  open_floating_prompt("Chuchu shell help", function(text)
    if text == "" then
      vim.notify("Chuchu: empty query", vim.log.levels.WARN)
      return
    end

    local cmd = { "chu", "chat" }
    local output = {}

    local job = vim.fn.jobstart(cmd, {
      stdout_buffered = true,
      on_stdout = function(_, data, _)
        if data then vim.list_extend(output, data) end
      end,
      on_exit = function()
        local raw = table.concat(output, "\n")
        vim.notify(raw, vim.log.levels.INFO)
      end,
      stdin = "pipe",
    })

    if job <= 0 then
      vim.notify("Chuchu: failed to start chat command", vim.log.levels.ERROR)
      return
    end

    vim.fn.chansend(job, text .. "\n")
    vim.fn.chanclose(job, "stdin")
  end)
end


local function detect_language()
  -- 1) Filetype heuristics
  local ft = vim.bo.filetype

  if ft == "elixir" or ft == "eelixir" then
    return "elixir"
  end

  if ft == "ruby" or ft == "eruby" then
    return "ruby"
  end

  if ft == "go" then
    return "go"
  end

  -- TypeScript/JavaScript filetypes → map para "ts"
  if ft == "typescript"
    or ft == "typescriptreact"
    or ft == "ts"
    or ft == "javascript"
    or ft == "javascriptreact"
    or ft == "jsx"
    or ft == "tsx" then
    return "ts"
  end

  -- 2) Project files in cwd
  local cwd = vim.fn.getcwd()

  -- Elixir
  if vim.fn.filereadable(cwd .. "/mix.exs") == 1 then
    return "elixir"
  end

  -- Ruby / Rails
  if vim.fn.filereadable(cwd .. "/Gemfile") == 1
    or vim.fn.filereadable(cwd .. "/config/application.rb") == 1 then
    return "ruby"
  end

  -- Go
  if vim.fn.filereadable(cwd .. "/go.mod") == 1 then
    return "go"
  end

  -- TypeScript/Node
  if vim.fn.filereadable(cwd .. "/tsconfig.json") == 1
    or vim.fn.filereadable(cwd .. "/package.json") == 1 then
    -- We treat TS/JS as TS for purposes of feature generation.
    return "ts"
  end

  return nil
end


local function open_floating_prompt(title, on_submit)
  local buf = vim.api.nvim_create_buf(false, true)
  local width = math.floor(vim.o.columns * 0.7)
  local height = 8
  local row = math.floor((vim.o.lines - height) / 2)
  local col = math.floor((vim.o.columns - width) / 2)

  local win = vim.api.nvim_open_win(buf, true, {
    relative = "editor",
    row = row,
    col = col,
    width = width,
    height = height,
    style = "minimal",
    border = "rounded",
    title = title,
    title_pos = "center",
  })

  vim.api.nvim_buf_set_lines(buf, 0, -1, false, {
    "Describe your feature. Chuchu will ask questions via the CLI.",
    "",
    "> ",
  })
  vim.api.nvim_win_set_cursor(win, { 3, 3 })
  vim.bo[buf].buftype = "prompt"
  vim.fn.prompt_setprompt(buf, "> ")

  vim.keymap.set("i", "<CR>", function()
    local lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
    local last = lines[#lines] or ""
    local input = last:gsub("^> ", "")
    vim.api.nvim_win_close(win, true)
    if on_submit then
      on_submit(input)
    end
  end, { buffer = buf })

  return buf, win
end

local function extract_all_blocks(text)
  local blocks = {}
  for tests_block, impl_block in text:gmatch("```tests(.-)```.-```impl(.-)```") do
    table.insert(blocks, { tests = tests_block, impl = impl_block })
  end
  return blocks
end

local function extract_lines(block)
  local lines = {}
  if not block or type(block) ~= "string" then return lines end
  for line in string.gmatch(block, "([^\n]*)\n?") do
    table.insert(lines, line)
  end
  return lines
end


function M.start_code_conversation()
  local lang = detect_language()
  if not lang then
    vim.notify("Chuchu: could not detect project language (Elixir/Ruby/Go/TS).", vim.log.levels.WARN)
    return
  end

  local feature_cmd = config.feature_cmds[lang]
  if not feature_cmd then
    vim.notify("Chuchu: no feature_cmd configured for language: " .. lang, vim.log.levels.ERROR)
    return
  end

  open_floating_prompt("Chuchu feature (" .. lang .. ")", function(text)
    if text == "" then
      vim.notify("Chuchu: empty feature description", vim.log.levels.WARN)
      return
    end

    local cmd = vim.deepcopy(feature_cmd)
    local output = {}

    local job = vim.fn.jobstart(cmd, {
      stdout_buffered = true,
      on_stdout = function(_, data, _)
        if data then vim.list_extend(output, data) end
      end,
      on_exit = function()
        local raw = table.concat(output, "\n")
        M.open_layout(text, raw, lang)
      end,
      stdin = "pipe",
    })

    if job <= 0 then
      vim.notify("Chuchu: failed to start feature command (" .. table.concat(cmd, " ") .. ")", vim.log.levels.ERROR)
      return
    end

    vim.fn.chansend(job, text .. "\n")
    vim.fn.chanclose(job, "stdin")
  end)
end

function M.open_layout(prompt, raw, lang)
  local blocks = extract_all_blocks(raw)
  
  if #blocks == 0 then
    vim.notify("Chuchu: no ```tests/```impl blocks found", vim.log.levels.WARN)
    vim.notify("Raw output: " .. raw:sub(1, 200), vim.log.levels.INFO)
    return
  end

  local filetype = "plaintext"
  if lang == "elixir" then
    filetype = "elixir"
  elseif lang == "ruby" then
    filetype = "ruby"
  elseif lang == "go" then
    filetype = "go"
  elseif lang == "ts" then
    filetype = "typescript"
  end

  local convo = {}
  table.insert(convo, "# Chuchu (" .. (lang or "unknown") .. ")")
  table.insert(convo, "")
  table.insert(convo, "## Prompt")
  table.insert(convo, prompt)
  table.insert(convo, "")
  table.insert(convo, "## Output")
  for line in raw:gmatch("([^\n]*)\n?") do
    table.insert(convo, line)
  end

  for idx, block in ipairs(blocks) do
    if idx > 1 then
      vim.cmd("tabnew")
    end

    vim.cmd("vsplit")
    
    local test_buf = vim.api.nvim_create_buf(true, false)
    local test_lines = extract_lines(block.tests)
    vim.api.nvim_buf_set_lines(test_buf, 0, -1, false, test_lines)
    vim.api.nvim_win_set_buf(vim.api.nvim_get_current_win(), test_buf)
    vim.bo[test_buf].filetype = filetype

    vim.cmd("wincmd l")
    
    local impl_buf = vim.api.nvim_create_buf(true, false)
    local impl_lines = extract_lines(block.impl)
    vim.api.nvim_buf_set_lines(impl_buf, 0, -1, false, impl_lines)
    vim.api.nvim_win_set_buf(vim.api.nvim_get_current_win(), impl_buf)
    vim.bo[impl_buf].filetype = filetype

    vim.cmd("vsplit")
    local chat_win = vim.api.nvim_get_current_win()
    if not chat_state.buf then
      chat_state.buf = vim.api.nvim_create_buf(true, false)
      vim.api.nvim_buf_set_lines(chat_state.buf, 0, -1, false, convo)
      vim.bo[chat_state.buf].filetype = "markdown"
    end
    vim.api.nvim_win_set_buf(chat_win, chat_state.buf)
    chat_state.win = chat_win
    
    vim.cmd("wincmd h")
    vim.cmd("wincmd h")
  end
end

  local filetype = "plaintext"
  if lang == "elixir" then
    filetype = "elixir"
  elseif lang == "ruby" then
    filetype = "ruby"
  elseif lang == "go" then
    filetype = "go"
  elseif lang == "ts" then
    -- For TS/JS we use "typescript" filetype.
    filetype = "typescript"
  end

  if tests_block then
    local buf = vim.api.nvim_create_buf(true, false)
    local lines = extract_lines(tests_block)
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
    vim.api.nvim_win_set_buf(vim.api.nvim_get_current_win(), buf)
    vim.bo[buf].filetype = filetype
  end

  if impl_block then
    vim.cmd("wincmd j")
    local buf = vim.api.nvim_create_buf(true, false)
    local lines = extract_lines(impl_block)
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
    vim.api.nvim_win_set_buf(vim.api.nvim_get_current_win(), buf)
    vim.bo[buf].filetype = filetype
  end
end


local function ensure_memory_dir()
  local mem_path = config.memory_file
  local dir = vim.fn.fnamemodify(mem_path, ":h")
  if vim.fn.isdirectory(dir) == 0 then
    vim.fn.mkdir(dir, "p")
  end
  return mem_path
end

local function json_escape(str)
  str = str:gsub("\\", "\\\\")
  str = str:gsub("\"", "\\\"")
  str = str:gsub("\n", "\\n")
  return str
end

local function current_language_for_feedback()
  local lang = detect_language()
  if lang then return lang end
  local ft = vim.bo.filetype
  if ft and ft ~= "" then
    return ft
  end
  return "unknown"
end

function M.record_feedback(kind)
  local mem_path = ensure_memory_dir()

  local lang = current_language_for_feedback()
  local buf = vim.api.nvim_get_current_buf()
  local file = vim.api.nvim_buf_get_name(buf)
  if file == "" then file = "[NoName]" end

  local ts = os.date("!%Y-%m-%dT%H:%M:%SZ")
  local text = table.concat(vim.api.nvim_buf_get_lines(buf, 0, -1, false), "\n")

  local entry = string.format(
    '{"timestamp":"%s","kind":"%s","language":"%s","file":"%s","snippet":"%s"}\n',
    json_escape(ts),
    json_escape(kind),
    json_escape(lang),
    json_escape(file),
    json_escape(text:sub(1, 4000))
  )

  local fh, err = io.open(mem_path, "a")
  if not fh then
    vim.notify("Chuchu: failed to open memory file: " .. tostring(err), vim.log.levels.ERROR)
    return
  end
  fh:write(entry)
  fh:close()

  vim.notify("Chuchu: feedback '" .. kind .. "' recorded for " .. file, vim.log.levels.INFO)
end

return M
