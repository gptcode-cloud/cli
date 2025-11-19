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
  -- Use explicit path to ensure we get the latest binary with status updates
  chat_cmd = { vim.fn.expand("~/.local/bin/chu"), "chat" },
  keymaps = {
    chat          = "<leader>cd",
    verified      = "<leader>vf",
    failed        = "<leader>fr",
    shell_help    = "<leader>xs",
  },
  memory_file = vim.fn.expand("~/.chuchu/memories.jsonl"),
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
--     dir = "~/workspace/chuchu/neovim",
--     config = function()
--       require("chuchu").setup()
--     end,
--   }
function M.setup(opts)
  config = vim.tbl_deep_extend("force", config, opts or {})

  vim.api.nvim_create_user_command("ChuchuChat", function()
    M.toggle_chat()
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

  vim.api.nvim_create_user_command("ChuchuModels", function()
    M.switch_model()
  end, {})

  local km = config.keymaps
  if km.chat and km.chat ~= "" then
    vim.keymap.set("n", km.chat, ":ChuchuChat<CR>", {
      silent = true,
      noremap = true,
      desc = "Chuchu: toggle chat",
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
  vim.keymap.set("n", "<C-d>", ":ChuchuChat<CR>", {
    silent = true,
    noremap = true,
    desc = "Chuchu: toggle chat",
  })
  vim.keymap.set("n", "<C-v>", ":ChuchuVerified<CR>", {
    silent = true,
    noremap = true,
    desc = "Chuchu: verified code",
  })
  vim.keymap.set("n", "<C-r>", ":ChuchuFailed<CR>", {
    silent = true,
    noremap = true,
    desc = "Chuchu: rejected code",
  })
  vim.keymap.set("n", "<C-x>", ":ChuchuShell<CR>", {
    silent = true,
    noremap = true,
    desc = "Chuchu: generate shell",
  })
  vim.keymap.set("n", "<C-m>", ":ChuchuModels<CR>", {
    silent = true,
    noremap = true,
    desc = "Chuchu: switch model",
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
  local setup_path = vim.fn.expand("~/.chuchu/setup.yaml")
  if vim.fn.filereadable(setup_path) == 0 then
    chat_state.model = "not configured"
    chat_state.backend = "not configured"
    chat_state.agent_models = {}
    return
  end
  
  local lines = vim.fn.readfile(setup_path)
  local in_defaults = false
  local in_target_backend = false
  local in_agent_models = false
  local target_backend = nil
  
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
    end
  end
  
  if target_backend then
    for _, line in ipairs(lines) do
      if line:match("^%s%s%s%s" .. target_backend .. ":") then
        in_target_backend = true
        chat_state.agent_models = {}
      elseif line:match("^%s%s%s%s[a-z]") and in_target_backend then
        in_target_backend = false
        in_agent_models = false
      end
      
      if in_target_backend and line:match("^%s%s%s%s%s%s%s%sagent_models:") then
        in_agent_models = true
      elseif in_agent_models then
        if line:match("^%s%s%s%s%s%s%s%s%s%s%s%s") then
          local agent, model = line:match("^%s+([^:]+):%s*(.+)$")
          if agent and model then
            chat_state.agent_models[agent] = model
          end
        else
          in_agent_models = false
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
  table.insert(lines, "🐺 Chuchu")
  table.insert(lines, string.format("Backend: %s", backend_display))
  
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
    vim.notify("Chuchu: could not find message separator", vim.log.levels.ERROR)
    return
  end
  
  local emoji_line = lines[separator_idx]
  local user_msg = emoji_line:gsub("^👤 %| ", ""):gsub("^%s*(.-)%s*$", "%1")
  
  if user_msg == "" then
    vim.notify("Chuchu: empty message", vim.log.levels.WARN)
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
  local catalog_path = vim.fn.expand("~/.chuchu/models.json")
  
  if vim.fn.filereadable(catalog_path) == 0 then
    vim.notify("Model catalog not found. Run 'chu models update'", vim.log.levels.WARN)
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
  
  vim.ui.select(options, {
    prompt = "Select backend:",
  }, function(choice)
    if not choice then return end
    local idx = tonumber(choice:match("^(%d+)"))
    local selected_backend = backends[idx].name
    
    M.show_model_configuration_menu(selected_backend)
  end)
end

function M.show_model_configuration_menu(backend)
  local options = {
    "1. Configure default model",
    "2. Configure agent models (router/query/editor/research)",
  }
  
  vim.ui.select(options, {
    prompt = "Configuration type:",
  }, function(choice)
    if not choice then return end
    local idx = tonumber(choice:match("^(%d+)"))
    
    if idx == 1 then
      M.configure_default_model(backend)
    elseif idx == 2 then
      M.configure_agent_models(backend)
    end
  end)
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

function M.configure_agent_models(backend)
  vim.notify(string.format("\nConfiguring models for %s\n", backend), vim.log.levels.INFO)
  
  local agents = {"router", "query", "editor", "research"}
  local agent_models = {}
  
  local function configure_next_agent(idx)
    if idx > #agents then
      M.save_agent_models(backend, agent_models)
      return
    end
    
    local agent = agents[idx]
    M.get_models_for_backend(backend, function(models)
      if #models == 0 then
        configure_next_agent(idx + 1)
        return
      end
      
      if #models > 0 and models[1] then
        agent_models[agent] = models[1]
      end
      
      configure_next_agent(idx + 1)
    end, agent)
  end
  
  configure_next_agent(1)
end

function M.save_agent_models(backend, agent_models)
  local setup_path = vim.fn.expand("~/.chuchu/setup.yaml")
  local lines = vim.fn.readfile(setup_path)
  local new_lines = {}
  local in_target_backend = false
  local in_models_section = false
  local agent_models_written = false
  local in_backend_section = false
  local backend_exists = false
  local backend_section_added = false
  
  for i, line in ipairs(lines) do
    if line:match("^backend:") and not backend_section_added then
      in_backend_section = true
      backend_section_added = true
      table.insert(new_lines, line)
    elseif line:match("^backend:") then
      -- Skip duplicate backend: lines (don't add to new_lines)
    elseif line:match("^[a-z]") and in_backend_section and not line:match("^%s") then
      in_backend_section = false
      table.insert(new_lines, line)
    elseif line:match("^%s%s%s%s" .. backend .. ":") then
      backend_exists = true
      in_target_backend = true
      table.insert(new_lines, line)
    elseif line:match("^%s%s%s%s[a-z]") and in_target_backend then
      if not agent_models_written then
        table.insert(new_lines, "        agent_models:")
        for agent, model in pairs(agent_models) do
          table.insert(new_lines, "            " .. agent .. ": " .. model)
        end
        agent_models_written = true
      end
      in_target_backend = false
      in_models_section = false
      table.insert(new_lines, line)
    elseif in_target_backend then
      if line:match("^%s%s%s%s%s%s%s%smodels:") then
        in_models_section = true
        table.insert(new_lines, line)
      elseif line:match("^%s%s%s%s%s%s%s%sagent_models:") then
        in_models_section = false
        local j = i + 1
        while j <= #lines and lines[j]:match("^%s%s%s%s%s%s%s%s%s%s%s%s") do
          j = j + 1
        end
        table.insert(new_lines, "        agent_models:")
        for agent, model in pairs(agent_models) do
          table.insert(new_lines, "            " .. agent .. ": " .. model)
        end
        agent_models_written = true
        for k = j, #lines do
          if k > i then
            break
          end
        end
      else
        table.insert(new_lines, line)
      end
    else
      table.insert(new_lines, line)
    end
  end
  
  if in_target_backend and not agent_models_written then
    table.insert(new_lines, "        agent_models:")
    for agent, model in pairs(agent_models) do
      table.insert(new_lines, "            " .. agent .. ": " .. model)
    end
  end
  
  if not backend_exists then
    local backend_section_found = false
    for i, line in ipairs(new_lines) do
      if line:match("^backend:") then
        backend_section_found = true
        local url_map = {
          groq = "https://api.groq.com/openai/v1",
          openrouter = "https://openrouter.ai/api/v1",
          openai = "https://api.openai.com/v1",
          deepseek = "https://api.deepseek.com/v1",
        }
        table.insert(new_lines, i + 1, "    " .. backend .. ":")
        table.insert(new_lines, i + 2, "        type: openai")
        table.insert(new_lines, i + 3, "        base_url: " .. (url_map[backend] or "https://api.openai.com/v1"))
        table.insert(new_lines, i + 4, "        default_model: ")
        table.insert(new_lines, i + 5, "        models: {}")
        table.insert(new_lines, i + 6, "        agent_models:")
        local insert_idx = i + 7
        for agent, model in pairs(agent_models) do
          table.insert(new_lines, insert_idx, "            " .. agent .. ": " .. model)
          insert_idx = insert_idx + 1
        end
        break
      end
    end
  end
  
  vim.fn.writefile(new_lines, setup_path)
  vim.notify(string.format("\n✓ Agent models configured for %s\n", backend), vim.log.levels.INFO)
end

function M.get_models_for_backend(backend, callback, agent)
  local catalog_path = vim.fn.expand("~/.chuchu/models.json")
  
  if vim.fn.filereadable(catalog_path) == 0 then
    vim.notify("Model catalog not found. Run 'chu models update'", vim.log.levels.WARN)
    callback({})
    return
  end
  
  local catalog_json = vim.fn.readfile(catalog_path)
  local catalog_str = table.concat(catalog_json, "\n")
  local ok, catalog = pcall(vim.fn.json_decode, catalog_str)
  
  if not ok or not catalog then
    vim.notify("Failed to parse model catalog", vim.log.levels.ERROR)
    callback({})
    return
  end
  
  local backend_data = catalog[backend]
  if not backend_data or not backend_data.models then
    vim.notify("No models found for " .. backend, vim.log.levels.WARN)
    callback({})
    return
  end
  
  local models = backend_data.models
  
  if agent then
    local filtered = {}
    for _, model in ipairs(models) do
      for _, rec in ipairs(model.recommended_for or {}) do
        if rec == agent then
          table.insert(filtered, model)
          break
        end
      end
    end
    if #filtered > 0 then
      models = filtered
    end
  end
  
  M.show_model_picker(models, callback, agent, backend)
end

function M.show_model_picker(models, callback, agent, backend)
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
  
  vim.ui.input({
    prompt = header .. prompt_text,
  }, function(query)
    if query == nil then
      callback({})
      return
    end
    
    local filtered = models
    if query and query ~= "" then
      filtered = {}
      local query_lower = query:lower()
      local query_normalized = query_lower:gsub("[^a-z0-9]", "")
      
      for _, model in ipairs(models) do
        local name_lower = model.name:lower()
        local id_lower = model.id:lower()
        local name_normalized = name_lower:gsub("[^a-z0-9]", "")
        local id_normalized = id_lower:gsub("[^a-z0-9]", "")
        
        if string.find(name_lower, query_lower, 1, true) or 
           string.find(id_lower, query_lower, 1, true) or
           string.find(name_normalized, query_normalized, 1, true) or
           string.find(id_normalized, query_normalized, 1, true) or
           M.has_tag(model.tags, query_lower) then
          table.insert(filtered, model)
        end
      end
      
      if #filtered == 0 then
        vim.schedule(function()
          vim.notify("\nNo models match '" .. query .. "'\n", vim.log.levels.WARN)
          vim.defer_fn(function()
            M.show_model_picker(models, callback, agent, backend)
          end, 100)
        end)
        return
      end
    end
    
    local options = {}
    for i, model in ipairs(filtered) do
      local tags_str = table.concat(model.tags or {}, ", ")
      local install_indicator = ""
      if backend == "ollama" then
        install_indicator = model.installed and "✓ " or "⬇ "
      end
      local line = string.format("%d. %s%s (%s) [%s] ($%.2f/$%.2f)", 
        i, install_indicator, model.name, model.id, tags_str,
        model.pricing_prompt_per_m_tokens or 0,
        model.pricing_completion_per_m_tokens or 0)
      table.insert(options, line)
    end
    
    vim.ui.select(options, {
      prompt = (backend and ("[" .. backend .. "] ") or "") .. "Select model:",
    }, function(choice)
      if not choice then
        callback({})
        return
      end
      local idx = tonumber(choice:match("^(%d+)"))
      if idx then
        local selected_model = filtered[idx]
        
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
end

function M.prompt_ollama_install(model, callback)
  vim.ui.input({
    prompt = string.format("Model '%s' not installed. Install now? (Y/n): ", model.id),
    default = "Y"
  }, function(response)
    if not response or response == "" or response:lower() == "y" then
      vim.notify(string.format("Installing %s...", model.id), vim.log.levels.INFO)
      
      vim.fn.jobstart({"ollama", "pull", model.id}, {
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

function M.has_tag(tags, query)
  if not tags then return false end
  for _, tag in ipairs(tags) do
    if tag:lower():match(query) then
      return true
    end
  end
  return false
end

function M.update_defaults(backend, model)
  local setup_path = vim.fn.expand("~/.chuchu/setup.yaml")
  local lines = vim.fn.readfile(setup_path)
  local new_lines = {}
  local in_defaults = false
  
  for _, line in ipairs(lines) do
    if line:match("^defaults:") then
      in_defaults = true
      table.insert(new_lines, line)
    elseif line:match("^[a-z]") and not line:match("^%s") then
      in_defaults = false
      table.insert(new_lines, line)
    elseif in_defaults then
      if line:match("^%s+backend:") then
        table.insert(new_lines, "    backend: " .. backend)
      elseif line:match("^%s+model:") then
        table.insert(new_lines, "    model: " .. model)
      else
        table.insert(new_lines, line)
      end
    else
      table.insert(new_lines, line)
    end
  end
  
  vim.fn.writefile(new_lines, setup_path)
  
  chat_state.backend = backend
  chat_state.model = model
  M.render_chat()
  
  vim.notify("Switched to " .. backend .. "/" .. model, vim.log.levels.INFO)
end

function M.shell_help()
  vim.ui.input({ prompt = "Shell command help: " }, function(text)
    if not text or text == "" then
      vim.notify("Chuchu: empty query", vim.log.levels.WARN)
      return
    end

    local cmd = { vim.fn.expand("~/.local/bin/chu"), "chat" }
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
    vim.notify("Chuchu: could not detect project language (Elixir/Ruby/Go/TS).", vim.log.levels.WARN)
    return
  end

  chat_state.lang = lang
  chat_state.conversation = {}

  open_floating_prompt("Chuchu (" .. lang .. ")", function(text)
    if text == "" then
      vim.notify("Chuchu: empty query", vim.log.levels.WARN)
      return
    end

    table.insert(chat_state.conversation, "User: " .. text)
    M.send_to_llm(text)
  end)
end

function M.send_to_llm(user_input)
  local cmd = vim.deepcopy(config.chat_cmd)
  local assistant_response = ""
  chat_state.active_tools = {}
  chat_state.tool_outputs = {}
  
  if os.getenv("CHUCHU_DEBUG") == "1" then
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
  
  if os.getenv("CHUCHU_DEBUG") == "1" then
    print("DEBUG: messages count = " .. #messages)
    for i, m in ipairs(messages) do
      print(string.format("  [%d] role=%s, content=%s", i, m.role, m.content:sub(1, 50)))
    end
  end

  local history_json = vim.fn.json_encode({messages = messages})
  
  chat_state.completed_tools = {}
  table.insert(chat_state.conversation, "Assistant: ")
  local assistant_idx = #chat_state.conversation

  -- Determine project root (where .git resides) to ensure tools can find source files
  local git_dir = vim.fn.finddir('.git', '.;')
  local project_root = git_dir ~= '' and vim.fn.fnamemodify(git_dir, ':h') or vim.fn.getcwd()
  local job = vim.fn.jobstart(cmd, { cwd = project_root,
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

  if job <= 0 then
    vim.notify("Chuchu: failed to start command", vim.log.levels.ERROR)
    return
  end

  vim.fn.chansend(job, history_json .. "\n")
  vim.fn.chanclose(job, "stdin")
end

function M.handle_tool_event(event_json, assistant_idx)
  local ok, event = pcall(vim.fn.json_decode, event_json)
  if not ok then return end
  
  if event.type == "tool_start" then
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
  table.insert(convo_lines, "# Chuchu (" .. (chat_state.lang or "unknown") .. ")")
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
    vim.notify("Chuchu: failed to open memory file: " .. tostring(err), vim.log.levels.ERROR)
    return
  end
  fh:write(entry)
  fh:close()

  vim.notify("Chuchu: feedback '" .. kind .. "' recorded for " .. file, vim.log.levels.INFO)
end

return M
