local pickers = require "telescope.pickers"
local finders = require "telescope.finders"
local actions = require "telescope.actions"
local actions_state = require "telescope.actions.state"
local entry_display = require "telescope.pickers.entry_display"
local conf = require("telescope.config").values

local function finder_fn()
    return function(prompt)
        local res = vim.g.spotify_search
        local results = {}

        for _, v in pairs(res) do
            table.insert(results, { v[1], v[2], v[3] })
        end

        return results
    end
end

local function entry_fn(opts)
    opts = opts or {}

    local display_items = {
        { width = 40 },
        { width = 25 },
    }

    local displayer = entry_display.create {
        separator = " by ",
        items = display_items
    }

    local make_display = function (entry)
        if vim.g.spotify_type == 'artists' or vim.g.spotify_type == 'playlists' then
            return displayer {
                { entry.track, "TelescopeResultsNumber" },
            }
        end
        
        return displayer {
            { entry.track, "TelescopeResultsNumber" },
            { entry.artist, "TelescopeResultsComment" },
        }
    end

    return function(entry)
        return {
            artist = entry[2],
            track = entry[1],
            uri = entry[3],
            display = make_display,
            ordinal = entry[1] .. entry[2],
        }
    end
end

local spotify = function (opts)
    opts = opts or {}
    pickers.new(opts, {
        prompt_title = "Spotify (" .. vim.g.spotify_type .. ": " .. vim.g.spotify_title .. ")",
        finder = finders.new_dynamic({
            entry_maker = entry_fn(opts),
            fn = finder_fn()
        }),
        sorter = conf.generic_sorter(opts),
        attach_mappings = function (prompt_bufnr, map)
            actions.select_default:replace(function()
                actions.close(prompt_bufnr)
                local selection = actions_state.get_selected_entry()
                local cmd = ":call SpotifyPlay('" .. selection.uri .. "')"
                vim.api.nvim_command(cmd)
            end)
            return true
        end
    }):find()
end

local list_devices = function (opts)
    opts = opts or {}
    pickers.new(opts, {
        prompt_title = "Connect to a Device",
        finder = finders.new_dynamic({
            entry_maker = function (entry)
                return {
                    value = entry,
                    display = entry[1],
                    ordinal = entry[1]
                }
            end,
            fn =  function(prompt)
                local res = vim.g.spotify_devices
                local results = {}

                for _, v in pairs(res) do
                    table.insert(results, { v[1] })
                end

                return results
            end
        }),
        sorter = conf.generic_sorter(opts),
        attach_mappings = function (prompt_bufnr, map)
            actions.select_default:replace(function()
                actions.close(prompt_bufnr)
                local selection = actions_state.get_selected_entry()
                vim.g.spotify_device = selection.value
            end)
            return true
        end
    }):find()
end

local M = {
    opts = {
        status = {
            update_interval = 10000,
            format = '%s %t by %a'
        }
    },
    status = {},
    _status_line = ""
}

M.namespace = 'Spotify'

function M.setup(opts)
    M.opts = vim.tbl_deep_extend("force", M.opts, opts)

	vim.api.nvim_set_keymap("n", "<Plug>(SpotifySkip)", ":<c-u>call SpotifyPlayback('next')<CR>", { silent = true })
	vim.api.nvim_set_keymap("n", "<Plug>(SpotifyPause)", ":<c-u>call SpotifyPlayback('pause')<CR>", { silent = true })
    vim.api.nvim_set_keymap("n", "<Plug>(SpotifySave)", ":<c-u>call SpotifySave()<CR>", { silent = true })
end

function M.init()
    spotify(require'telescope.themes'.get_dropdown{})
end

function M.devices()
    list_devices(require'telescope.themes'.get_dropdown{})
end

function M.status:start()
    local timer = vim.loop.new_timer()
   timer:start(1000, M.opts.status.update_interval, vim.schedule_wrap(function ()
        local cmd = "spt playback --status --format '" .. M.opts.status.format .. "'"
        vim.fn.jobstart(cmd, { on_stdout = self.on_event, stdout_buffered = true })
    end))
end

function M.status:on_event(data)
    if data then
        M._status_line = data[1]
    end
end

function M.status:listen()
    return M._status_line
end


return M
