> [!CAUTION]
> This project and package are no longer maintained and likely do not work.

<details>
  <summary>Original readme</summary>

# clnotify

clnotify is a golang application to notify on Craigslist posts matching specific terms to a Discord webhook.

## Usage

A configuration file is required:

`craigslist.area_id` can be found by searching: https://reference.craigslist.org/Areas
`search_distance` is in KM
`searches[].categories` can be found here: https://github.com/ecnepsnai/craigslist/blob/main/categories.md

For example, the following looks for "vintage" and "retro" in the computers for sale category of Vancouver's craigslist

```json
{
    "craigslist": {
        "area_id": 16,
        "latitude": 49.2810,
        "longitude": -123.0400,
        "search_distance": 30
    },
    "discord": {
        "webhook_url": "https://discord.com/api/webhooks/..."
    },
    "searches": [
        {
            "categories": [ "sya" ],
            "query": "retro",
            "name": "Retro Computers"
        },
        {
            "categories": [ "sya" ],
            "query": "vintage",
            "name": "Vintage Computers",
            "ignore": [
                "dell"
            ]
        }
    ]
}
```

</details>
