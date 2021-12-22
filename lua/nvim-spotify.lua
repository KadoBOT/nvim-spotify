local pickers = require "telescope.pickers"
local finders = require "telescope.finders"
local actions = require "telescope.actions"
local actions_state = require "telescope.actions.state"
local entry_display = require "telescope.pickers.entry_display"
local conf = require("telescope.config").values

local function finder_fn()
    return function(prompt)
        local res = vim.g.spotify_search
        local search_type = vim.g.spotify_type
        local results = {}

        if search_type == "artist" then
            for _, v in pairs(res.Artists.Artists) do
                table.insert(results, { v.Name, "", v.URI })
            end
        end

        if search_type == "playlist" then
            for _, v in pairs(res.Playlists.Playlists) do
                table.insert(results, { v.Name, v.Owner.DisplayName, v.URI })
            end
        end

        if search_type == "album" then
            for _, v in pairs(res.Albums.Albums) do
                local artist_name = ""
                for i, artist in ipairs(v.Artists) do
                    if i == 1 then
                        artist_name = artist_name ..  artist.Name
                    elseif not v.Artists[i + 1] then
                        artist_name = artist_name .. " and " .. artist.Name
                    else
                        artist_name = artist_name .. ", " .. artist.Name
                    end
                end
                table.insert(results, { v.Name, artist_name, v.URI })
            end
        end

        if search_type == "track" then
            for _, v in pairs(res.Tracks.Tracks) do
                local artist_name = ""
                for i, artist in ipairs(v.Artists) do
                    if i == 1 then
                        artist_name = artist_name ..  artist.Name
                    elseif not v.Artists[i + 1] then
                        artist_name = artist_name .. " and " .. artist.Name
                    else
                        artist_name = artist_name .. ", " .. artist.Name
                    end
                end
                table.insert(results, { v.Name, artist_name, v.URI })
            end
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
        if vim.g.spotify_type == 'artist' then
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

return {
    init = function ()
        spotify(require'telescope.themes'.get_dropdown{})
    end,
    setup = function (opts)
        vim.g.spotify_refresh_token = opts.refresh_token
    end
}
