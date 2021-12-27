# 🎵nvim-spotify

For productivity addicts who enjoy coding while listening to Spotify, and cannot lose their focus switching to the app to control their music.

`nvim-spotify` requires [spotify-tui](https://github.com/Rigellute/spotify-tui)

## Features

- Display/Filter the search results with Telescope
- Currently playing statusline.
- Pause/Resume a track
- Skip a track
- Add a track to the library
- Display the name of what's being played
- Select which device to play on
- Search by:
  - Track (`<C-T> or CR`)
  - Album (`<C-L>`)
  - Playlist (`<C-Y>`)
  - Artist (`<C-R>`)

## Requirements
> `nvim-spotify` is a wrapper for `spotify-tui`, therefore, it is required for this plugin to work. Check [their Github
> repository for installation instructions](https://github.com/Rigellute/spotify-tui#installation)

- [Spotify TUI](https://github.com/Rigellute/spotify-tui)
- Golang
- Telescope

## Installation

### [packer](https://github.com/wbthomason/packer.nvim)
```lua
-- Lua
use {
    'KadoBOT/nvim-spotify', 
    requires = 'nvim-telescope/telescope.nvim',
    config = function()
        local spotify = require'nvim-spotify'

        spotify.setup {
            -- default opts
            status = {
                update_interval = 10000, -- the interval (ms) to check for what's currently playing
                format = '%s %t by %a' -- spotify-tui --format argument
            }
        }
    end,
    run = 'make'
}
```

### [vim-plug](https://github.com/junegunn/vim-plug)
```viml
Plug 'KadoBOT/nvim-spotify', { 'do': 'make' }
```

#### Notes
Decreasing the `update_interval` value means more API calls in a shorter period. Because of the Spotify API rate limiter, setting this too low can block future requests.
Besides that, those constant updates can make your computer slow. 
**So bear this in mind when changing this value.**

## Usage
`nvim-spotify` has two commands:

### Connecting to a Device
Use this command to select which device Spotify should play on.
```
:SpotifyDevices
```

### Opening search input
Spotify Search input. Check the keymaps below for Search shortcuts.
```
:Spotify
```

### Default keymaps:
The following keymaps are set by default when the Spotify search input is open:
| mode | key | Description |
|---|---|---|
| normal | Esc | Close
| normal | q | Close
| normal, insert | C-T | Search for Tracks
| normal, insert | C-Y | Search for Playlists
| normal, insert | C-L | Search for Albums
| normal, insert | C-R | Search for Artists

### Extra keymaps
 You can also define the additional following keymaps
```lua
vim.api.nvim_set_keymap("n", "<leader>sn", "<Plug>(SpotifySkip)",  { silent = true }) -- Skip the current track
vim.api.nvim_set_keymap("n", "<leader>sp", "<Plug>(SpotifyPause)", , { silent = true }) -- Pause/Resume the current track
vim.api.nvim_set_keymap("n", "<leader>ss", "<Plug>(SpotifySave)",  { silent = true }) -- Add the current track to your library
vim.api.nvim_set_keymap("n", "<leader>so", ":Spotify<CR>",  { silent = true }) -- Open Spotify Search window
vim.api.nvim_set_keymap("n", "<leader>sd", ":SpotifyDevices<CR>",  { silent = true }) -- Open Spotify Devices window
```

### Statusline
You can display what's currently playing on your statusline. The example below shows how to show it on [lualine](https://github.com/nvim-lualine/lualine.nvim),
although the configuration should be quite similar on other statusline plugins:
```lua
local status = require'nvim-spotify'.status

status:start()

require('lualine').setup {
    sections = {
        lualine_x = {
            status.listen
        }
    }
}
```

