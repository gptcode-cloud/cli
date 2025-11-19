vim.o.loadplugins = true
vim.cmd([[set runtimepath+=~/.local/share/nvim/lazy/lazy.nvim]])
vim.cmd([[set runtimepath+=~/workspace/opensource/chuchu/neovim]])

local chuchu = require('chuchu')
chuchu.setup({})

vim.schedule(function()
  print("[TEST] Starting headless test")
  
  local test_job_id = 999
  local stdout_callback = nil
  local exit_callback = nil
  local stdin_data = nil
  
  vim.fn.jobstart = function(cmd, opts)
    print("[TEST] jobstart called")
    stdout_callback = opts.on_stdout
    exit_callback = opts.on_exit
    
    vim.schedule(function()
      local events = {
        '__EVENT__{"type":"message","data":{"content":"🚀 Starting guided workflow..."}}__EVENT__',
        '__EVENT__{"type":"status","data":{"status":"🔄 Analyzing task..."}}__EVENT__',
        '__EVENT__{"type":"open_plan","data":{"path":"/tmp/test_draft.md"}}__EVENT__',
        '__EVENT__{"type":"message","data":{"content":"📋 Draft plan created. Review in opened tab."}}__EVENT__',
        '__EVENT__{"type":"confirm","data":{"id":"plan_draft","prompt":"Proceed with detailed planning?"}}__EVENT__',
      }
      
      for _, event in ipairs(events) do
        print("[TEST] Sending event:", event:sub(1, 80))
        if stdout_callback then
          stdout_callback(test_job_id, {event}, nil)
        end
        vim.wait(100)
      end
      
      print("[TEST] Waiting for user input...")
      vim.wait(2000)
      
      if stdout_callback then
        stdout_callback(test_job_id, {'__EVENT__{"type":"message","data":{"content":"❌ Cancelled by user"}}__EVENT__'}, nil)
      end
      
      vim.wait(100)
      
      if exit_callback then
        exit_callback(test_job_id, 0, nil)
      end
    end)
    
    return test_job_id
  end
  
  vim.fn.chansend = function(job, data)
    print("[TEST] chansend called with job:", job)
    stdin_data = data
    print("[TEST] Received stdin:", data:sub(1, 150))
    return 1
  end
  
  local original_input = vim.ui.input
  vim.ui.input = function(opts, callback)
    print("[TEST] vim.ui.input called with prompt:", opts.prompt)
    vim.schedule(function()
      print("[TEST] Simulating user response: y")
      callback("y")
    end)
  end
  
  chuchu.toggle_chat()
  
  vim.schedule(function()
    vim.wait(200)
    print("[TEST] Simulating user message")
    
    local buf = vim.api.nvim_get_current_buf()
    local lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
    print("[TEST] Buffer has", #lines, "lines")
    
    vim.api.nvim_buf_set_option(buf, "modifiable", true)
    vim.api.nvim_buf_set_lines(buf, -1, -1, false, {"👤 | Analyse this codebase and add pix payment"})
    
    chuchu.send_message_from_buffer()
    
    vim.schedule(function()
      vim.wait(5000)
      print("[TEST] Test complete, exiting")
      vim.cmd('qall!')
    end)
  end)
end)

vim.schedule(function()
  vim.wait(8000)
  print("[TEST] Timeout reached, forcing exit")
  vim.cmd('qall!')
end)
