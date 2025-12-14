-- gptcode.nvim – generic Neovim integration for GPTCode
--
-- Features:
-- - Detects project type (Elixir / Ruby / Go / TypeScript) via simple heuristics.
-- - Uses corresponding CLI commands:
--     Elixir     -> `gptcode feature-elixir`
--     Ruby       -> `gptcode feature-ruby`   (you implement this in Go)
--     Go         -> `gptcode feature-go`     (you implement this in Go)
--     TypeScript -> `gptcode feature-ts`     (already scaffolded)
-- - :GPTCodeFeature → opens a floating prompt for the feature description.
-- - Renders ```tests / ```impl fenced blocks from stdout and opens a 3-pane layout:
--     left top: tests
--     left bottom: implementation
--     right: conversation (prompt + raw output)
-- - Feedback commands store snapshots in ~/.gptcode/memories.jsonl:
--     :GPTCodeFeedbackGood  (default key: <leader>ck)
--     :GPTCodeFeedbackBad   (default key: <leader>cx)

local M = {}

local config = {
  chat_cmd = { "gptcode", "chat" },
  keymaps = {
    chat          = "<leader>cd",
    verified      = "<leader>vf",
    failed        = "<leader>fr",
    shell_help    = "<leader>xs",
  },
  memory_file = vim.fn.expand("~/.gptcode/memories.jsonl"),
}

local chat_state = { 
  buf = nil, 
  win = nil,
  conversation = {},
  lang = nil,
  job = nil,
  model = nil,
  backend = nil,
  agent_models = {},
  tools_buf = nil,
  tools_win = nil,
  active_tools = {},
  completed_tools = {},
  tool_outputs = {}
}

local function detect_language()
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

  if ft == "typescript"
    or ft == "typescriptreact"
    or ft == "ts"
    or ft == "javascript"
    or ft == "javascriptreact"
    or ft == "jsx"
    or ft == "tsx" then
    return "ts"
  end

  local cwd = vim.fn.getcwd()

  if vim.fn.filereadable(cwd .. "/mix.exs") == 1 then
    return "elixir"
  end

  if vim.fn.filereadable(cwd .. "/Gemfile") == 1
    or vim.fn.filereadable(cwd .. "/config/application.rb") == 1 then
    return "ruby"
  end

  if vim.fn.filereadable(cwd .. "/go.mod") == 1 then
    return "go"
  end

  if vim.fn.filereadable(cwd .. "/tsconfig.json") == 1
    or vim.fn.filereadable(cwd .. "/package.json") == 1 then
    return "ts"
  end

  return nil
end

--- Setup to be called from your plugin manager.
-- Example (lazy.nvim):
--   {
--     dir = "~/workspace/gptcode/neovim",
--     config = function()
--       require("gptcode").setup()
--     end,
--   }
function M.setup(opts)
  config = vim.tbl_deep_extend("force", config, opts or {})

  vim.api.nvim_create_user_command("GPTCodeChat", function()
    M.toggle_chat()
  end, {})

  vim.api.nvim_create_user_command("GPTCodeVerified", function()
    M.record_feedback("good")
  end, {})

  vim.api.nvim_create_user_command("GPTCodeFailed", function()
    M.record_feedback("bad")
  end, {})

  vim.api.nvim_create_user_command("GPTCodeShell", function()
    M.shell_help()
  end, {})

  vim.api.nvim_create_user_command("GPTCodeModels", function()
    M.switch_model()
  end, {})

  vim.api.nvim_create_user_command("GPTCodeAuto", function(opts)
    local default = vim.fn.getcwd() .. "/docs/plans/maestro-autonomous-execution-plan.md"
    vim.ui.input({ prompt = "Plan file:", default = default }, function(input)
      if not input or input == "" then return end
      local cmd = { "gptcode", "implement", input, "--auto" }
      vim.fn.jobstart(cmd, {
        env = vim.tbl_extend("force", vim.fn.environ(), {GPTCODE_NVIM_MODE = "1"}),
        stdout_buffered = true,
        on_stdout = function(_, data)
          if not data then return end
          for _, line in ipairs(data) do
            if line ~= "" then
              vim.schedule(function()
                vim.notify(line, vim.log.levels.INFO, { title = "GPTCode Implement" })
              end)
            end
          end
        end,
        on_stderr = function(_, data)
          if not data then return end
          for _, line in ipairs(data) do
            if line ~= "" then
              vim.schedule(function()
                vim.notify(line, vim.log.levels.WARN, { title = "GPTCode Implement" })
              end)
            end
          end
        end,
      })
    end)
  end, { nargs = "?" })

  vim.api.nvim_create_user_command("GPTCodeModelSearch", function()
    M.search_models()
  end, {})

  local km = config.keymaps
  if km.chat and km.chat ~= "" then
    vim.keymap.set("n", km.chat, ":GPTCodeChat<CR>", {
      silent = true,
      noremap = true,
      desc = "GPTCode: toggle chat",
    })
  end
  if km.verified and km.verified ~= "" then
    vim.keymap.set("n", km.verified, ":GPTCodeVerified<CR>", {
      silent = true,
      noremap = true,
      desc = "GPTCode: verified code",
    })
  end
  if km.failed and km.failed ~= "" then
    vim.keymap.set("n", km.failed, ":GPTCodeFailed<CR>", {
      silent = true,
      noremap = true,
      desc = "GPTCode: failed code",
    })
  end
  if km.shell_help and km.shell_help ~= "" then
    vim.keymap.set("n", km.shell_help, ":GPTCodeShell<CR>", {
      silent = true,
      noremap = true,
      desc = "GPTCode: shell help",
    })
  end
  vim.keymap.set("n", "<C-d>", ":GPTCodeChat<CR>", {
    silent = true,
    noremap = true,
    desc = "GPTCode: toggle chat",
  })
  vim.keymap.set("n", "<leader>ca", ":GPTCodeAuto<CR>", {
    silent = true,
    noremap = true,
    desc = "GPTCode: autonomous implement",
  })
  vim.keymap.set("n", "<C-v>", ":GPTCodeVerified<CR>", {
    silent = true,
    noremap = true,
    desc = "GPTCode: verified code",
  })
  vim.keymap.set("n", "<C-r>", ":GPTCodeFailed<CR>", {
    silent = true,
    noremap = true,
    desc = "GPTCode: rejected code",
  })
  vim.keymap.set("n", "<C-x>", ":GPTCodeShell<CR>", {
    silent = true,
    noremap = true,
    desc = "GPTCode: generate shell",
  })
  vim.keymap.set("n", "<C-m>", ":GPTCodeModels<CR>", {
    silent = true,
    noremap = true,
    desc = "GPTCode: switch model",
  })
  vim.keymap.set("n", "<leader>ms", ":GPTCodeModelSearch<CR>", {
    silent = true,
    noremap = true,
    desc = "GPTCode: search models",
  })
  
  vim.api.nvim_create_autocmd("VimLeave", {
    callback = function()
      chat_state.conversation = {}
      chat_state.completed_tools = {}
      chat_state.active_tools = {}
    end,
  })
end

function M.toggle_chat()
  if chat_state.win and vim.api.nvim_win_is_valid(chat_state.win) then
    vim.api.nvim_win_close(chat_state.win, true)
    chat_state.win = nil
    chat_state.conversation = {}
    chat_state.completed_tools = {}
    chat_state.active_tools = {}
    return
  end

  if not chat_state.lang then
    chat_state.lang = detect_language() or "unknown"
  end

  if not chat_state.buf or not vim.api.nvim_buf_is_valid(chat_state.buf) then
    chat_state.buf = vim.api.nvim_create_buf(true, false)
    vim.bo[chat_state.buf].filetype = "markdown"
    
    M.render_chat()
    
    vim.api.nvim_buf_set_keymap(chat_state.buf, "n", "<CR>", "", {
      callback = function()
        M.send_message_from_buffer()
      end,
      noremap = true,
      silent = true,
    })
  end

  vim.cmd("vsplit")
  chat_state.win = vim.api.nvim_get_current_win()
  vim.api.nvim_win_set_buf(chat_state.win, chat_state.buf)
  vim.api.nvim_win_set_width(chat_state.win, 60)
  
  local line_count = vim.api.nvim_buf_line_count(chat_state.buf)
  vim.api.nvim_win_set_cursor(chat_state.win, {line_count, 0})
  
  vim.cmd("startinsert!")
end

function M.load_profile_info()
  local setup_path = vim.fn.expand("~/.gptcode/setup.yaml")
  if vim.fn.filereadable(setup_path) == 0 then
    chat_state.model = "not configured"
    chat_state.backend = "not configured"
    chat_state.profile = "default"
    chat_state.agent_models = {}
    return
  end
  
  local lines = vim.fn.readfile(setup_path)
  local in_defaults = false
  local target_backend = nil
  local target_profile = "default"
  
  for _, line in ipairs(lines) do
    if line:match("^defaults:") then
      in_defaults = true
    elseif line:match("^[a-z]") and not line:match("^%s") then
      in_defaults = false
    end
    
    if in_defaults then
      local backend_match = line:match("^%s+backend:%s*(.+)$")
      if backend_match then
        chat_state.backend = backend_match
        target_backend = backend_match
      end
      local model_match = line:match("^%s+model:%s*(.+)$")
      if model_match then
        chat_state.model = model_match
      end
      local profile_match = line:match("^%s+profile:%s*(.+)$")
      if profile_match then
        target_profile = profile_match
        chat_state.profile = profile_match
      end
    end
  end
  
  if target_backend then
    local in_target_backend = false
    local in_profiles = false
    local in_target_profile = false
    local in_agent_models = false
    local in_backend_agent_models = false
    
    chat_state.agent_models = {}
    
    for _, line in ipairs(lines) do
      if line:match("^%s%s%s%s" .. target_backend .. ":") then
        in_target_backend = true
      elseif line:match("^%s%s%s%s[a-z]") and in_target_backend then
        in_target_backend = false
        in_profiles = false
        in_target_profile = false
        in_agent_models = false
        in_backend_agent_models = false
      end
      
      if in_target_backend and line:match("^%s%s%s%s%s%s%s%sprofiles:") then
        in_profiles = true
        in_backend_agent_models = false
      elseif in_target_backend and line:match("^%s%s%s%s%s%s%s%sagent_models:") and not in_profiles then
        in_backend_agent_models = true
      elseif in_profiles and line:match("^%s%s%s%s%s%s%s%s%s%s%s%s" .. target_profile .. ":") then
        in_target_profile = true
        in_backend_agent_models = false
      elseif in_target_profile and line:match("^%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%sagent_models:") then
        in_agent_models = true
      elseif in_agent_models then
        if line:match("^%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s") then
          local agent, model = line:match("^%s+([^:]+):%s*(.+)$")
          if agent and model then
            chat_state.agent_models[agent] = model
          end
        else
          in_agent_models = false
        end
      elseif in_backend_agent_models and target_profile == "default" then
        if line:match("^%s%s%s%s%s%s%s%s%s%s%s%s") and not line:match("^%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s") then
          local agent, model = line:match("^%s+([^:]+):%s*(.+)$")
          if agent and model then
            chat_state.agent_models[agent] = model
          end
        else
          in_backend_agent_models = false
        end
      end
    end
  end
  
  if not chat_state.model then
    chat_state.model = "default"
  end
  if not chat_state.backend then
    chat_state.backend = "ollama"
  end
  if not chat_state.profile then
    chat_state.profile = "default"
  end
  if not chat_state.agent_models then
    chat_state.agent_models = {}
  end
end

function M.render_chat()
  if not chat_state.buf or not vim.api.nvim_buf_is_valid(chat_state.buf) then
    return
  end
  
  if not chat_state.model then
    M.load_profile_info()
  end
  
  local lines = {}
  local cwd = vim.fn.getcwd()
  local repo_name = vim.fn.fnamemodify(cwd, ":t")
  
  local backend_display = chat_state.backend and chat_state.backend:sub(1,1):upper()..chat_state.backend:sub(2) or "?"
  local profile_display = chat_state.profile or "default"
  table.insert(lines, "🐺 GPTCode")
  table.insert(lines, string.format("Backend: %s / %s", backend_display, profile_display))
  
  if chat_state.agent_models and next(chat_state.agent_models) then
    for agent, model in pairs(chat_state.agent_models) do
      local short_model = model:match("([^/]+)$") or model
      if #short_model > 30 then
        short_model = short_model:sub(1, 27) .. "..."
      end
      table.insert(lines, string.format("  %s: %s", agent, short_model))
    end
  elseif chat_state.model then
    table.insert(lines, string.format("Model: %s", chat_state.model))
  end
  
  table.insert(lines, "[Esc + Enter send | ^D close | ^M models | ^X shell]")
  table.insert(lines, string.rep("-", 60))
  
  if #chat_state.conversation == 0 then
    table.insert(lines, "")
  else
    for _, msg in ipairs(chat_state.conversation) do
      if msg:match("^User: ") then
        local content = msg:gsub("^User: ", "")
        for line in content:gmatch("([^\n]*)\n?") do
          if line ~= "" then
            table.insert(lines, "👤 | " .. line)
          end
        end
      elseif msg:match("^Assistant: ") then
        local content = msg:gsub("^Assistant: ", "")
        for line in content:gmatch("([^\n]*)\n?") do
          if line ~= "" then
            table.insert(lines, "🐺 | " .. line)
          end
        end
      end
    end
  end
  
  table.insert(lines, "👤 | ")
  
  vim.api.nvim_buf_set_option(chat_state.buf, "modifiable", true)
  vim.api.nvim_buf_set_lines(chat_state.buf, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(chat_state.buf, "modified", false)
end

function M.send_message_from_buffer()
  if not chat_state.buf or not vim.api.nvim_buf_is_valid(chat_state.buf) then
    return
  end
  
  local lines = vim.api.nvim_buf_get_lines(chat_state.buf, 0, -1, false)
  local separator_idx = nil
  for i = #lines, 1, -1 do
    local line = lines[i]
    if line:match("^👤 %|") then
      separator_idx = i
      break
    end
  end
  
  if not separator_idx then
    vim.notify("GPTCode: could not find message separator", vim.log.levels.ERROR)
    return
  end
  
  local emoji_line = lines[separator_idx]
  local user_msg = emoji_line:gsub("^👤 %| ", ""):gsub("^%s*(.-)%s*$", "%1")
  
  if user_msg == "" then
    vim.notify("GPTCode: empty message", vim.log.levels.WARN)
    return
  end
  
  table.insert(chat_state.conversation, "User: " .. user_msg)
  M.render_chat()
  M.show_loading_animation()
  M.send_to_llm(user_msg)
end

function M.show_loading_animation()
  if not chat_state.win or not vim.api.nvim_win_is_valid(chat_state.win) then
    return
  end
  
  local frames = {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
  local frame_idx = 1
  
  local timer = vim.loop.new_timer()
  chat_state.loading_timer = timer
  chat_state.current_status = "Thinking..."
  
  timer:start(0, 100, vim.schedule_wrap(function()
    if not chat_state.buf or not vim.api.nvim_buf_is_valid(chat_state.buf) then
      timer:stop()
      return
    end
    
    local lines = vim.api.nvim_buf_get_lines(chat_state.buf, 0, -1, false)
    local last_line = #lines
    
    vim.api.nvim_buf_set_option(chat_state.buf, "modifiable", true)
    vim.api.nvim_buf_set_lines(chat_state.buf, last_line - 1, last_line, false, 
      {frames[frame_idx] .. " | " .. chat_state.current_status})
    vim.api.nvim_buf_set_option(chat_state.buf, "modifiable", false)
    
    frame_idx = (frame_idx % #frames) + 1
  end))
end

function M.stop_loading_animation()
  if chat_state.loading_timer then
    chat_state.loading_timer:stop()
    chat_state.loading_timer = nil
  end
end

function M.switch_model()
  local catalog_path = vim.fn.expand("~/.gptcode/models.json")
  
  if vim.fn.filereadable(catalog_path) == 0 then
    vim.notify("Model catalog not found. Run 'gptcode models update'", vim.log.levels.WARN)
    return
  end
  
  local catalog_json = vim.fn.readfile(catalog_path)
  local catalog_str = table.concat(catalog_json, "\n")
  local ok, catalog = pcall(vim.fn.json_decode, catalog_str)
  
  if not ok or not catalog then
    vim.notify("Failed to parse model catalog", vim.log.levels.ERROR)
    return
  end
  
  local backends = {}
  for backend, data in pairs(catalog) do
    if data.models and #data.models > 0 then
      table.insert(backends, {name = backend, count = #data.models})
    end
  end
  
  table.sort(backends, function(a, b) return a.name < b.name end)
  
  if #backends == 0 then
    vim.notify("No backends found in catalog", vim.log.levels.WARN)
    return
  end
  
  local options = {}
  for i, backend in ipairs(backends) do
    table.insert(options, string.format("%d. %s (%d models)", i, backend.name, backend.count))
  end
  
  print("")
  vim.ui.select(options, {
    prompt = "Select backend:",
  }, function(choice)
    if not choice then return end
    local idx = tonumber(choice:match("^(%d+)"))
    local selected_backend = backends[idx].name
    
    print("")
    M.show_model_configuration_menu(selected_backend)
  end)
end

function M.show_model_configuration_menu(backend)
  local profiles = M.list_profiles(backend)
  local options = {}
  
  table.insert(options, "→ Create new profile")
  table.insert(options, "")
  
  for _, profile_name in ipairs(profiles) do
    table.insert(options, string.format("  %s", profile_name))
  end
  
  print("")
  vim.ui.select(options, {
    prompt = string.format("[%s] Profile Management:", backend),
  }, function(choice)
    if not choice then return end
    
    if choice:match("^→") then
      M.create_profile_interactive(backend)
    elseif choice ~= "" then
      local profile_name = choice:match("^%s+(.+)$")
      M.show_profile_actions(backend, profile_name)
    end
  end)
end

function M.list_profiles(backend)
  local output = vim.fn.system({"gptcode", "profiles", "list", backend})
  
  if vim.v.shell_error ~= 0 then
    return {"default"}
  end
  
  local ok, profiles = pcall(vim.fn.json_decode, output)
  if not ok or not profiles or #profiles == 0 then
    return {"default"}
  end
  
  return profiles
end

function M.create_profile_interactive(backend)
  vim.ui.input({
    prompt = string.format("[%s] New profile name: ", backend)
  }, function(name)
    if not name or name == "" then
      vim.notify("Profile creation cancelled", vim.log.levels.WARN)
      return
    end
    
    if name == "default" then
      vim.notify("Cannot create profile named 'default'", vim.log.levels.ERROR)
      return
    end
    
    local result = vim.fn.system({"gptcode", "profiles", "create", backend, name})
    if vim.v.shell_error == 0 then
      vim.notify(string.format("✓ Created profile: %s/%s", backend, name), vim.log.levels.INFO)
      M.configure_profile_agents(backend, name)
    else
      vim.notify("Failed to create profile: " .. result, vim.log.levels.ERROR)
    end
  end)
end

function M.show_profile_actions(backend, profile_name)
  local actions = {
    "1. Load profile",
    "2. Configure agents",
    "3. Show details",
    "4. Delete profile (if not default)"
  }
  
  vim.ui.select(actions, {
    prompt = string.format("[%s/%s]", backend, profile_name)
  }, function(choice)
    if not choice then return end
    
    if choice:match("^1") then
      M.load_profile(backend, profile_name)
    elseif choice:match("^2") then
      M.configure_profile_agents(backend, profile_name)
    elseif choice:match("^3") then
      M.show_profile_details(backend, profile_name)
    elseif choice:match("^4") then
      M.delete_profile(backend, profile_name)
    end
  end)
end

function M.configure_profile_agents(backend, profile_name)
  local agents = {"router", "query", "editor", "research"}
  local agent_idx = 1
  
  local function configure_next_agent()
    if agent_idx > #agents then
      vim.notify(string.format("✓ Profile %s/%s configured", backend, profile_name), vim.log.levels.INFO)
      M.load_profile(backend, profile_name)
      return
    end
    
    local agent = agents[agent_idx]
    M.get_models_for_backend(backend, function(models)
      if #models == 0 then
        agent_idx = agent_idx + 1
        configure_next_agent()
        return
      end
      
      local selected_model = models[1]
      if selected_model then
        vim.fn.system({"gptcode", "profiles", "set-agent", backend, profile_name, agent, selected_model})
        agent_idx = agent_idx + 1
        configure_next_agent()
      end
    end, agent)
  end
  
  configure_next_agent()
end

function M.show_profile_details(backend, profile_name)
  local output = vim.fn.system({"gptcode", "profiles", "show", backend, profile_name})
  
  if vim.v.shell_error == 0 then
    local lines = vim.split(output, "\n")
    local formatted = {}
    table.insert(formatted, string.format("Profile: %s/%s", backend, profile_name))
    table.insert(formatted, "")
    
    for _, line in ipairs(lines) do
      if line ~= "" and not line:match("^Profile:") then
        table.insert(formatted, line)
      end
    end
    
    vim.notify(table.concat(formatted, "\n"), vim.log.levels.INFO)
  else
    vim.notify("Failed to show profile details", vim.log.levels.ERROR)
  end
end

function M.delete_profile(backend, profile_name)
  if profile_name == "default" then
    vim.notify("Cannot delete default profile", vim.log.levels.ERROR)
    return
  end
  
  vim.ui.input({
    prompt = string.format("Delete profile %s/%s? (y/N): ", backend, profile_name),
    default = "N"
  }, function(response)
    if response and response:lower() == "y" then
      local result = vim.fn.system({"gptcode", "profiles", "delete", backend, profile_name})
      if vim.v.shell_error == 0 then
        vim.notify(string.format("✓ Deleted profile: %s/%s", backend, profile_name), vim.log.levels.INFO)
      else
        vim.notify("Failed to delete profile: " .. result, vim.log.levels.ERROR)
      end
    else
      vim.notify("Deletion cancelled", vim.log.levels.INFO)
    end
  end)
end

function M.load_profile(backend, profile_name)
  vim.fn.system({"gptcode", "config", "set", "defaults.backend", backend})
  if vim.v.shell_error ~= 0 then
    vim.notify("Failed to set backend", vim.log.levels.ERROR)
    return
  end
  
  vim.fn.system({"gptcode", "config", "set", "defaults.profile", profile_name})
  if vim.v.shell_error ~= 0 then
    vim.notify("Failed to set profile", vim.log.levels.ERROR)
    return
  end
  
  M.load_profile_info()
  M.render_chat()
  
  vim.notify(string.format("Loaded profile: %s/%s", backend, profile_name), vim.log.levels.INFO)
end

function M.configure_default_model(backend)
  M.get_models_for_backend(backend, function(models)
    if #models == 0 then
      vim.notify("No models found for " .. backend, vim.log.levels.WARN)
      return
    end
    
    local model_options = {}
    for i, model in ipairs(models) do
      table.insert(model_options, i .. ". " .. model)
    end
    
    vim.ui.select(model_options, {
      prompt = "Select default model:",
    }, function(model_choice)
      if not model_choice then return end
      local model_idx = tonumber(model_choice:match("^(%d+)"))
      local selected_model = models[model_idx]
      
      M.update_defaults(backend, selected_model)
    end)
  end)
end


function M.get_models_for_backend(backend, callback, agent)
  M.show_model_picker(callback, agent, backend)
end

function M.show_model_picker(callback, agent, backend)
  local agent_descriptions = {
    router = "Router (fast intent classification)",
    query = "Query (read/analyze code)",
    editor = "Editor (write/modify code)",
    research = "Research (web search/docs)"
  }
  
  local prompt_text = "Filter models (or Enter for all): "
  if agent and agent_descriptions[agent] then
    prompt_text = string.format("%s - filter or Enter: ", agent_descriptions[agent])
  elseif agent then
    prompt_text = string.format("Select model for %s (filter or Enter): ", agent)
  end
  
  local header = ""
  if backend then
    header = string.format("[Backend: %s]\n\n", backend)
  end
  
  print("")
  vim.ui.input({
    prompt = header .. prompt_text,
  }, function(query)
    if query == nil then
      callback({})
      return
    end
    
    query = query or ""
    
    local cmd = {"gptcode", "models", "search"}
    if query ~= "" then
      for term in query:gmatch("%S+") do
        table.insert(cmd, term)
      end
    end
    table.insert(cmd, "--backend")
    table.insert(cmd, backend)
    if agent and agent ~= "" then
      table.insert(cmd, "--agent")
      table.insert(cmd, agent)
    end
    
    local stdout_data = {}
    vim.fn.jobstart(cmd, {
      env = vim.tbl_extend("force", vim.fn.environ(), {GPTCODE_NVIM_MODE = "1"}),
      stdout_buffered = true,
      on_stdout = function(_, data)
        if data then
          stdout_data = data
        end
      end,
      on_exit = function(_, exit_code)
        if exit_code ~= 0 then
          vim.schedule(function()
            vim.notify("Failed to search models", vim.log.levels.ERROR)
            callback({})
          end)
          return
        end
        
        local json_str = table.concat(stdout_data, "\n")
        local ok, models = pcall(vim.fn.json_decode, json_str)
        
        if not ok or not models or #models == 0 then
          vim.schedule(function()
            vim.notify("\nNo models match '" .. query .. "'\n", vim.log.levels.WARN)
            vim.defer_fn(function()
              M.show_model_picker(callback, agent, backend)
            end, 100)
          end)
          return
        end
        
        local options = {}
        for i, model in ipairs(models) do
          local install_indicator = ""
          if backend == "ollama" then
            install_indicator = model.installed and "✓ " or "⬇ "
          end
          
          local recommended_indicator = ""
          if model.recommended then
            recommended_indicator = " ✓"
          end
          
          local name_display = model.name
          if backend == "ollama" then
            name_display = install_indicator .. name_display
          end
          
          local name_with_rec = name_display .. recommended_indicator
          local padding = string.rep(" ", math.max(1, 50 - #name_with_rec))
          
          local price_in = model.pricing_prompt_per_m_tokens or 0
          local price_out = model.pricing_completion_per_m_tokens or 0
          local ctx = model.context_window or 0
          local ctx_str = ctx > 0 and string.format("%dk", math.floor(ctx / 1000)) or ""
          
          local price_display = price_in == 0 and price_out == 0 
            and "[FREE]" 
            or string.format("$%.2f/$%.2f", price_in, price_out)
          
          local line = string.format("%d. %s%s%s %s",
            i, name_with_rec, padding, price_display,
            ctx_str ~= "" and string.format("(%s)", ctx_str) or "")
          table.insert(options, line)
        end
        
        vim.schedule(function()
          local prompt_msg = (backend and ("[" .. backend .. "] ") or "") .. "Select model:\nPrice per 1M tokens (Input/Output) - ✓ Role-based recommendation"
          vim.ui.select(options, {
            prompt = prompt_msg,
          }, function(choice)
            if not choice then
              callback({})
              return
            end
            local idx = tonumber(choice:match("^(%d+)"))
            if idx then
              local selected_model = models[idx]
              
              if backend == "ollama" and not selected_model.installed then
                M.prompt_ollama_install(selected_model, function(success)
                  if success then
                    callback({selected_model.id})
                  else
                    callback({})
                  end
                end)
              else
                callback({selected_model.id})
              end
            else
              callback({})
            end
          end)
        end)
      end,
    })
  end)
end

function M.prompt_ollama_install(model, callback)
  vim.ui.input({
    prompt = string.format("Model '%s' not installed. Install now? (Y/n): ", model.id),
    default = "Y"
  }, function(response)
    if not response or response == "" or response:lower() == "y" then
      vim.notify(string.format("Installing %s...", model.id), vim.log.levels.INFO)
      
      vim.fn.jobstart({"gptcode", "models", "install", model.id}, {
        env = vim.tbl_extend("force", vim.fn.environ(), {GPTCODE_NVIM_MODE = "1"}),
        on_exit = function(_, exit_code)
          if exit_code == 0 then
            vim.schedule(function()
              vim.notify(string.format("✓ %s installed successfully", model.id), vim.log.levels.INFO)
              callback(true)
            end)
          else
            vim.schedule(function()
              vim.notify(string.format("✗ Failed to install %s", model.id), vim.log.levels.ERROR)
              callback(false)
            end)
          end
        end,
        on_stdout = function(_, data)
          if data then
            vim.schedule(function()
              for _, line in ipairs(data) do
                if line ~= "" then
                  print(line)
                end
              end
            end)
          end
        end
      })
    else
      vim.notify("Installation cancelled", vim.log.levels.WARN)
      callback(false)
    end
  end)
end

function M.search_models()
  vim.ui.input({
    prompt = "Search models (e.g., 'ollama llama3', 'groq coding fast'): "
  }, function(query)
    if not query or query == "" then
      vim.notify("Search cancelled", vim.log.levels.WARN)
      return
    end
    
    local terms = {}
    for term in query:gmatch("%S+") do
      table.insert(terms, term)
    end
    
    local cmd = {"gptcode", "models", "search"}
    for _, term in ipairs(terms) do
      table.insert(cmd, term)
    end
    
    local output = {}
    vim.fn.jobstart(cmd, {
      env = vim.tbl_extend("force", vim.fn.environ(), {GPTCODE_NVIM_MODE = "1"}),
      stdout_buffered = true,
      on_stdout = function(_, data)
        if data then
          vim.list_extend(output, data)
        end
      end,
      on_exit = function(_, exit_code)
        if exit_code ~= 0 then
          vim.schedule(function()
            vim.notify("Failed to search models", vim.log.levels.ERROR)
          end)
          return
        end
        
        local json_str = table.concat(output, "\n")
        local ok, models = pcall(vim.fn.json_decode, json_str)
        
        if not ok or not models or #models == 0 then
          vim.schedule(function()
            vim.notify("No models found", vim.log.levels.WARN)
          end)
          return
        end
        
        vim.schedule(function()
          M.show_search_results(models, query)
        end)
      end
    })
  end)
end

function M.show_search_results(models, query)
  local options = {}
  
  for i, model in ipairs(models) do
    local tags_str = table.concat(model.tags or {}, ", ")
    local price_in = model.pricing_prompt_per_m_tokens or 0
    local price_out = model.pricing_completion_per_m_tokens or 0
    local ctx = model.context_window or 0
    local installed_mark = model.installed and " ✓" or ""
    
    local price_display = price_in == 0 and price_out == 0
      and "[FREE]"
      or string.format("$%.2f/$%.2f", price_in, price_out)
    
    local ctx_str = ctx > 0 and string.format("%dk", math.floor(ctx / 1000)) or ""
    
    local line = string.format("%d. %s%s [%s] %s %s",
      i, model.name, installed_mark, tags_str, price_display, ctx_str)
    table.insert(options, line)
  end
  
  vim.ui.select(options, {
    prompt = string.format("Search: '%s' (%d results)", query, #models)
  }, function(choice)
    if not choice then return end
    
    local idx = tonumber(choice:match("^(%d+)"))
    if not idx then return end
    
    local selected = models[idx]
    M.handle_model_selection(selected)
  end)
end

function M.handle_model_selection(model)
  local backend = model.id:match("^([^/]+)/") or "unknown"
  local is_ollama = backend == "ollama" or not model.id:match("/")
  
  if is_ollama and not model.installed then
    M.prompt_ollama_install(model, function(success)
      if success then
        M.show_model_actions(model)
      end
    end)
  else
    M.show_model_actions(model)
  end
end

function M.show_model_actions(model)
  local actions = {
    "1. Set as default model",
    "2. Use for current session",
    "3. Cancel"
  }
  
  vim.ui.select(actions, {
    prompt = string.format("Model: %s", model.name)
  }, function(choice)
    if not choice then return end
    
    if choice:match("^1") then
      local backend = model.id:match("^([^/]+)/") or "ollama"
      vim.fn.system({"gptcode", "config", "set", "defaults.model", model.id})
      vim.fn.system({"gptcode", "config", "set", "defaults.backend", backend})
      M.load_profile_info()
      M.render_chat()
      vim.notify(string.format("✓ Set as default: %s", model.id), vim.log.levels.INFO)
    elseif choice:match("^2") then
      chat_state.model = model.id
      M.render_chat()
      vim.notify(string.format("✓ Using for session: %s", model.id), vim.log.levels.INFO)
    end
  end)
end


function M.update_defaults(backend, model)
  vim.fn.system({"gptcode", "config", "set", "defaults.backend", backend})
  if vim.v.shell_error ~= 0 then
    vim.notify("Failed to set backend", vim.log.levels.ERROR)
    return
  end
  
  vim.fn.system({"gptcode", "config", "set", "defaults.model", model})
  if vim.v.shell_error ~= 0 then
    vim.notify("Failed to set model", vim.log.levels.ERROR)
    return
  end
  
  chat_state.backend = backend
  chat_state.model = model
  M.render_chat()
  
  vim.notify("Switched to " .. backend .. "/" .. model, vim.log.levels.INFO)
end

function M.shell_help()
  vim.ui.input({ prompt = "Shell command help: " }, function(text)
    if not text or text == "" then
      vim.notify("GPTCode: empty query", vim.log.levels.WARN)
      return
    end

    local cmd = { "gptcode", "chat" }
    local output = {}

    local job = vim.fn.jobstart(cmd, {
      env = vim.tbl_extend("force", vim.fn.environ(), {GPTCODE_NVIM_MODE = "1"}),
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
      vim.notify("GPTCode: failed to start chat command", vim.log.levels.ERROR)
      return
    end

    vim.fn.chansend(job, text .. "\n")
    vim.fn.chanclose(job, "stdin")
  end)
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
    "Describe your feature. GPTCode will ask questions via the CLI.",
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
  -- Capture any code block: ```lang ... ```
  for lang, content in text:gmatch("```(%w+)(.-)```") do
    -- Skip if it's just a small inline snippet or empty
    if #content > 10 then
      table.insert(blocks, { lang = lang, content = content })
    end
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
    vim.notify("GPTCode: could not detect project language (Elixir/Ruby/Go/TS).", vim.log.levels.WARN)
    return
  end

  chat_state.lang = lang
  chat_state.conversation = {}

  open_floating_prompt("GPTCode (" .. lang .. ")", function(text)
    if text == "" then
      vim.notify("GPTCode: empty query", vim.log.levels.WARN)
      return
    end

    table.insert(chat_state.conversation, "User: " .. text)
    M.send_to_llm(text)
  end)
end

function M.handle_confirmation(prompt, id)
  vim.schedule(function()
    vim.ui.input({
      prompt = prompt .. " [y/n]: ",
      default = "y"
    }, function(response)
      if not response then response = "n" end
      response = response:lower()
      
      if chat_state.job and chat_state.job > 0 then
        local answer = (response == "y" or response == "yes") and "y\n" or "n\n"
        vim.fn.chansend(chat_state.job, answer)
      end
    end)
  end)
end

function M.send_to_llm(user_input)
  local cmd = vim.deepcopy(config.chat_cmd)
  local assistant_response = ""
  chat_state.active_tools = {}
  chat_state.tool_outputs = {}
  
  local event_file = vim.fn.expand("~/.gptcode/events.jsonl")
  vim.fn.writefile({}, event_file)
  vim.notify("[DEBUG] Events file cleared: " .. event_file, vim.log.levels.INFO)
  
  local watch_timer = vim.loop.new_timer()
  local last_size = 0
  local events_processed = 0
  watch_timer:start(0, 100, vim.schedule_wrap(function()
    local size = vim.fn.getfsize(event_file)
    if size > last_size then
      last_size = size
      local lines = vim.fn.readfile(event_file)
      if #lines > events_processed then
        for i = events_processed + 1, #lines do
          local event_json = lines[i]
          local ok, event = pcall(vim.fn.json_decode, event_json)
          if ok and event then
            if os.getenv("GPTCODE_DEBUG") == "1" then
              print("[WATCHER] Processing event #" .. i .. ": " .. event.type)
            end
            M.handle_tool_event(event_json, #chat_state.conversation)
            events_processed = i
          end
        end
      end
    end
  end))
  
  if os.getenv("GPTCODE_DEBUG") == "1" then
    print("DEBUG: conversation size = " .. #chat_state.conversation)
    for i, c in ipairs(chat_state.conversation) do
      print(string.format("  [%d] %s", i, c:sub(1, 60)))
    end
  end
  
  local messages = {}
  for _, msg in ipairs(chat_state.conversation) do
    if msg:match("^User: ") then
      table.insert(messages, {role = "user", content = msg:sub(7)})
    elseif msg:match("^Assistant: ") then
      local content = msg:sub(13)
      
      if content:find("\226") or content:find("\240") then
      else
        content = content:gsub("%s+$", "")
        if content ~= "" and content ~= "[No response]" then
          table.insert(messages, {role = "assistant", content = content})
        end
      end
    end
  end
  
  if os.getenv("GPTCODE_DEBUG") == "1" then
    print("DEBUG: messages count = " .. #messages)
    for i, m in ipairs(messages) do
      print(string.format("  [%d] role=%s, content=%s", i, m.role, m.content:sub(1, 50)))
    end
  end

  local history_json = vim.fn.json_encode({messages = messages})
  
  chat_state.completed_tools = {}
  table.insert(chat_state.conversation, "Assistant: ")
  local assistant_idx = #chat_state.conversation

  local git_dir = vim.fn.finddir('.git', '.;')
  local project_root = git_dir ~= '' and vim.fn.fnamemodify(git_dir, ':h') or vim.fn.getcwd()
  chat_state.job = vim.fn.jobstart(cmd, { 
    env = vim.tbl_extend("force", vim.fn.environ(), {GPTCODE_NVIM_MODE = "1"}),
    cwd = project_root,
    stdout_buffered = false,
    on_stdout = function(_, data, _)
      if not data then return end
      
      for _, line in ipairs(data) do
        if line ~= "" then
          local event_match = line:match("__EVENT__(.+)__EVENT__")
          if event_match then
            M.handle_tool_event(event_match, assistant_idx)
          else
            if assistant_response ~= "" and line ~= "" then
              assistant_response = assistant_response .. " "
            end
            assistant_response = assistant_response .. line
            
            if #chat_state.completed_tools == 0 then
              chat_state.conversation[assistant_idx] = "Assistant: " .. assistant_response
              M.render_chat()
            end
          end
        end
      end
    end,
    on_stderr = function(_, data, _)
      if not data then return end
      for _, line in ipairs(data) do
        if line ~= "" then
          local status_match = line:match("%[STATUS%]%s*(.+)")
          if status_match then
            chat_state.current_status = status_match
          end
        end
      end
    end,
    on_exit = function()
      M.stop_loading_animation()
      if watch_timer then
        watch_timer:stop()
        watch_timer:close()
      end
      if assistant_response == "" then
        chat_state.conversation[assistant_idx] = "Assistant: [No response]"
      else
        chat_state.conversation[assistant_idx] = "Assistant: " .. assistant_response .. " ✓"
      end
      M.render_chat()
      
      vim.schedule(function()
        if not chat_state.win or not vim.api.nvim_win_is_valid(chat_state.win) then return end
        local line_count = vim.api.nvim_buf_line_count(chat_state.buf)
        vim.api.nvim_win_set_cursor(chat_state.win, {line_count, 0})
        vim.cmd("startinsert!")
      end)
    end,
    stdin = "pipe",
  })

  if chat_state.job <= 0 then
    vim.notify("GPTCode: failed to start command", vim.log.levels.ERROR)
    return
  end

  vim.notify("[DEBUG] Job started: " .. chat_state.job, vim.log.levels.INFO)
  
  vim.schedule(function()
    vim.fn.chansend(chat_state.job, history_json .. "\n")
    vim.fn.chanclose(chat_state.job, 'stdin')
    vim.notify("[DEBUG] Stdin sent and closed, watcher active", vim.log.levels.INFO)
  end)
end

function M.handle_tool_event(event_json, assistant_idx)
  if os.getenv("GPTCODE_DEBUG") == "1" then
    print("[PLUGIN] Event: " .. event_json:sub(1, 100) .. " type=" .. (vim.fn.json_decode(event_json).type or "?"))
  end
  
  local ok, event = pcall(vim.fn.json_decode, event_json)
  if not ok then return end
  
  if event.type == "message" then
    local msg = event.data and event.data.content or ""
    table.insert(chat_state.conversation, "Assistant: " .. msg)
    M.render_chat()
    
  elseif event.type == "status" then
    local status = event.data and event.data.status or ""
    chat_state.current_status = status
    vim.schedule(function()
      vim.notify(status, vim.log.levels.INFO)
    end)
    
  elseif event.type == "confirm" then
    local prompt = event.data and event.data.prompt or "Proceed?"
    local id = event.data and event.data.id or "confirm"
    M.handle_confirmation(prompt, id)
    
  elseif event.type == "open_plan" then
    local path = event.data and event.data.path or ""
    if path ~= "" then
      vim.schedule(function()
        vim.cmd("wincmd l")
        vim.cmd("edit " .. vim.fn.fnameescape(path))
        vim.cmd("wincmd h")
      end)
    end
    
  elseif event.type == "open_split" then
    local test_file = event.data and event.data.test_file or ""
    local impl_file = event.data and event.data.impl_file or ""
    if test_file ~= "" and impl_file ~= "" then
      vim.schedule(function()
        vim.cmd("tabnew")
        vim.cmd("edit " .. vim.fn.fnameescape(test_file))
        vim.cmd("vsplit")
        vim.cmd("wincmd l")
        vim.cmd("edit " .. vim.fn.fnameescape(impl_file))
        vim.cmd("wincmd h")
      end)
    end
    
  elseif event.type == "complete" then
    table.insert(chat_state.conversation, "Assistant: ✅ Complete")
    M.render_chat()
    
  elseif event.type == "notify" then
    local message = event.data and event.data.message or ""
    local level = event.data and event.data.level or "info"
    local vim_level = vim.log.levels.INFO
    
    if level == "warn" then
      vim_level = vim.log.levels.WARN
    elseif level == "error" then
      vim_level = vim.log.levels.ERROR
    end
    
    vim.schedule(function()
      vim.notify(message, vim_level)
    end)
    
  elseif event.type == "tool_start" then
    table.insert(chat_state.active_tools, event.tool)
    M.ensure_tools_window()
    M.append_tool_output("\n⚙ " .. event.tool .. "\n")
    
    local current = chat_state.conversation[assistant_idx] or "Assistant: "
    local base = current:match("^(Assistant:.-)%s*[⚙✓]") or current
    base = base:gsub("%s+$", "")
    chat_state.conversation[assistant_idx] = base .. " ⚙ " .. event.tool
    M.render_chat()
    
  elseif event.type == "tool_end" then
    for i, tool in ipairs(chat_state.active_tools) do
      if tool == event.tool then
        table.remove(chat_state.active_tools, i)
        break
      end
    end
    
    if event.error then
      M.append_tool_output("\n❌ Error: " .. event.error .. "\n")
      table.insert(chat_state.completed_tools, "✗ " .. event.tool)
    else
      local output = event.result or ""
      if #output > 500 then
        output = output:sub(1, 500) .. "\n... (truncated, see tool window)"
      end
      M.append_tool_output("\n" .. output .. "\n")
      table.insert(chat_state.completed_tools, "✓ " .. event.tool)
      
      if event.tool == "write_file" and event.path then
        vim.schedule(function()
          local full_path = event.path
          if not vim.startswith(full_path, "/") then
            full_path = vim.fn.getcwd() .. "/" .. full_path
          end
          
          if chat_state.tools_win and vim.api.nvim_win_is_valid(chat_state.tools_win) then
            vim.api.nvim_set_current_win(chat_state.tools_win)
            vim.cmd("edit " .. vim.fn.fnameescape(full_path))
          end
        end)
      end
    end
    
    local current = chat_state.conversation[assistant_idx] or "Assistant: "
    local base = current:match("^(Assistant:.-)%s*[⚙✓✗]") or current
    base = base:gsub("%s+$", "")
    
    local status = ""
    if #chat_state.active_tools > 0 then
      status = " ⚙ " .. chat_state.active_tools[1]
    elseif #chat_state.completed_tools > 0 then
      status = " " .. table.concat(chat_state.completed_tools, " ")
    end
    
    chat_state.conversation[assistant_idx] = base .. status
    M.render_chat()
  end
end

function M.ensure_tools_window()
  if not chat_state.tools_buf or not vim.api.nvim_buf_is_valid(chat_state.tools_buf) then
    chat_state.tools_buf = vim.api.nvim_create_buf(false, true)
    vim.bo[chat_state.tools_buf].filetype = "markdown"
    vim.api.nvim_buf_set_lines(chat_state.tools_buf, 0, -1, false, {"# Tool Execution", ""})
  end
  
  if not chat_state.tools_win or not vim.api.nvim_win_is_valid(chat_state.tools_win) then
    local all_wins = vim.api.nvim_list_wins()
    local main_win = nil
    
    for _, win in ipairs(all_wins) do
      if win ~= chat_state.win and vim.api.nvim_win_is_valid(win) then
        local buf = vim.api.nvim_win_get_buf(win)
        if vim.api.nvim_buf_get_option(buf, 'buftype') == '' then
          main_win = win
          break
        end
      end
    end
    
    if main_win then
      chat_state.tools_win = main_win
      vim.api.nvim_win_set_buf(main_win, chat_state.tools_buf)
    else
      vim.cmd("split")
      chat_state.tools_win = vim.api.nvim_get_current_win()
      vim.api.nvim_win_set_buf(chat_state.tools_win, chat_state.tools_buf)
    end
  end
end

function M.append_tool_output(text)
  if not chat_state.tools_buf or not vim.api.nvim_buf_is_valid(chat_state.tools_buf) then
    return
  end
  
  local lines = vim.split(text, "\n")
  local current_lines = vim.api.nvim_buf_get_lines(chat_state.tools_buf, 0, -1, false)
  
  for _, line in ipairs(lines) do
    table.insert(current_lines, line)
  end
  
  vim.api.nvim_buf_set_lines(chat_state.tools_buf, 0, -1, false, current_lines)
  
  if chat_state.tools_win and vim.api.nvim_win_is_valid(chat_state.tools_win) then
    local line_count = vim.api.nvim_buf_line_count(chat_state.tools_buf)
    vim.api.nvim_win_set_cursor(chat_state.tools_win, {line_count, 0})
  end
end

function M.handle_llm_response(raw)
  M.stop_loading_animation()
  
  table.insert(chat_state.conversation, "Assistant: " .. raw)
  
  M.render_chat()
  
  local blocks = extract_all_blocks(raw)
  if #blocks > 0 then
    -- Create a new tab for the code blocks
    vim.cmd("tabnew")
    local combined_content = ""
    
    for _, block in ipairs(blocks) do
      combined_content = combined_content .. "// Language: " .. block.lang .. "\n" .. block.content .. "\n\n"
    end

    local buf = vim.api.nvim_create_buf(false, true)
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, vim.split(combined_content, "\n"))
    vim.api.nvim_set_current_buf(buf)
    vim.bo.filetype = blocks[1].lang -- Set filetype to the first block's language
  end
end

function M.show_chat_panel()
  if not chat_state.buf or not vim.api.nvim_buf_is_valid(chat_state.buf) then
    chat_state.buf = vim.api.nvim_create_buf(true, false)
    vim.bo[chat_state.buf].filetype = "markdown"
  end

  local convo_lines = {}
  table.insert(convo_lines, "# GPTCode (" .. (chat_state.lang or "unknown") .. ")")
  table.insert(convo_lines, "")
  
  for _, line in ipairs(chat_state.conversation) do
    for sub_line in line:gmatch("([^\n]*)\n?") do
      table.insert(convo_lines, sub_line)
    end
    table.insert(convo_lines, "")
  end
  
  table.insert(convo_lines, "")
  table.insert(convo_lines, "---")
  table.insert(convo_lines, "Press 'i' and type to continue the conversation, then <CR> to send")
  
  vim.api.nvim_buf_set_lines(chat_state.buf, 0, -1, false, convo_lines)
  
  if not chat_state.win or not vim.api.nvim_win_is_valid(chat_state.win) then
    vim.cmd("vsplit")
    chat_state.win = vim.api.nvim_get_current_win()
    vim.api.nvim_win_set_buf(chat_state.win, chat_state.buf)
    vim.api.nvim_win_set_width(chat_state.win, 60)
  end
  
  vim.api.nvim_buf_set_keymap(chat_state.buf, "n", "<CR>", "", {
    callback = function()
      M.continue_conversation()
    end,
    noremap = true,
    silent = true,
  })
end

function M.continue_conversation()
  open_floating_prompt("Continue conversation", function(text)
    if text == "" then return end
    table.insert(chat_state.conversation, "User: " .. text)
    M.send_to_llm(text)
  end)
end

function M.create_code_tabs(blocks)
  local filetype = "plaintext"
  if chat_state.lang == "elixir" then
    filetype = "elixir"
  elseif chat_state.lang == "ruby" then
    filetype = "ruby"
  elseif chat_state.lang == "go" then
    filetype = "go"
  elseif chat_state.lang == "ts" then
    filetype = "typescript"
  end

  for idx, block in ipairs(blocks) do
    vim.cmd("tabnew")

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
    vim.api.nvim_win_set_buf(chat_win, chat_state.buf)
    vim.api.nvim_win_set_width(chat_win, 60)
    
    vim.cmd("wincmd h")
    vim.cmd("wincmd h")
  end
  
  vim.notify("Created " .. #blocks .. " code tab(s)", vim.log.levels.INFO)
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
    vim.notify("GPTCode: failed to open memory file: " .. tostring(err), vim.log.levels.ERROR)
    return
  end
  fh:write(entry)
  fh:close()

  vim.notify("GPTCode: feedback '" .. kind .. "' recorded for " .. file, vim.log.levels.INFO)
end

return M
