if exists('g:loaded_gptcode') || !has('nvim')
  finish
endif
let g:loaded_gptcode = 1

lua require('gptcode').setup()
