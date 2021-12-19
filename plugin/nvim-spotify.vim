if exists('g:loaded_nvim_spotify')
    finish
endif

let g:loaded_nvim_spotify = 1

function! s:RequireNvimSpotify(host) abort
    return jobstart(['nvim-spotify'], {'rpc': v:true})
endfunction

call remote#host#Register('nvim-spotify', 'x', function('s:RequireNvimSpotify'))

call remote#host#RegisterPlugin('nvim-spotify', '0', [
    \ {'type': 'command', 'name': 'Spotify', 'sync': 0, 'opts': {}},
    \ {'type': 'function', 'name': 'SpotifySearch', 'sync': 0, 'opts': {}},
    \ {'type': 'function', 'name': 'SpotifyPlay', 'sync': 0, 'opts': {}},
    \ {'type': 'function', 'name': 'SpotifyCloseWin', 'sync': 0, 'opts': {}},
    \ ])

" vim:ts=4:sw=4:et