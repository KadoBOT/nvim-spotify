local pickers = require "telescope.pickers"
local finders = require "telescope.finders"
local entry_display = require "telescope.pickers.entry_display"
local conf = require("telescope.config").values

local function finder_fn()
    return function(prompt)
        local res = vim.g.spotify_search
        local results = {}

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
            table.insert(results, { v.Name, artist_name })
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
        return displayer {
            { entry.track, "TelescopeResultsNumber" },
            { entry.artist, "TelescopeResultsComment" },
        }
    end

    return function(entry)
        return {
            artist = entry[2],
            track = entry[1],
            display = make_display,
            ordinal = entry[1] .. entry[2],
        }
    end
end

local colors = function (opts)
    opts = opts or {}
    pickers.new(opts, {
        prompt_title = "Spotify (" .. vim.g.spotify_title .. ")",
        finder = finders.new_dynamic({
            entry_maker = entry_fn(opts),
            fn = finder_fn()
        }),
        sorter = conf.generic_sorter(opts)
    }):find()
end

return {
    init = function ()
        colors(require'telescope.themes'.get_dropdown{})
    end
}
