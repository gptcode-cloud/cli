if exists('g:loaded_chuchu') || !has('nvim')
  finish
endif
let g:loaded_chuchu = 1

lua require('chuchu').setup()
