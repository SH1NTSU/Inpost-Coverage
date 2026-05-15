
# [InPost Network Control]

## Author

- **Name:** [Marcel Budziszewski]
- **Email:** [u84_marbud_waw@technischools.com]

## Overview

[In 2–3 sentences, describe what you built. What does it do? What problem does it solve or what question does it answer?] I live in the village under Warsaw so the problem for me was the distance to the closest InPost point. **InPost Network Control** shows where is the best location to place a new point according to the nearest POI as well as including competition and habitability across Poland.

## Demo & Description

[Describe your solution in detail. What does it do? How does it work? What approach did you take and why? Cover the key technical choices, architecture, and anything else that helps us understand your project without reading every line of code.] 

**InPost Network Control** is a tool for finding the best places to install new
InPost points across Poland.

The backend pulls existing locker positions from the provided API into
Postgres. For each ~1.5 km cell on the map it asks two questions — *how far is
the nearest InPost?* and *how far is the nearest competitor?* — producing four
tiers: Greenfield (no operator within 1 km), Competitive gap (only a
competitor nearby), InPost only, and
Saturated.

Raw distance isn't enough since most of Poland is farmland. A cell only counts
as inhabited if there's an OSM signal nearby: a commercial POI within 1.5 km,
a town node within 1.5 km, or a village node within 500 m. Water, forest, and
farmland are excluded via OSM polygons. Hamlet nodes are ignored — they tag
5-house clusters and would paint whole regions red.

Recommendations run the same scan but try to snap each underserved spot to a
real shop or fuel station within 250 m. About 70% are anchored ("install a
Paczkomat at this Żabka"); the rest are open areas where a private host could
provide a site. Competitor and POI data come from the Overpass API — ~13k
competitor lockers and ~180k anchor POIs total.

The biggest issue was latency. A naive implementation paints
23 000+ rectangles per province and the browser melts. Three things keep it
usable:

1. Server-side cell cap — top 1 500 greenfield + 500 competitive cells per
   province. Stats stay accurate; render set stays small.
2. Cached per-province snapshots in Redis + Postgres, warmed in parallel
   on startup so most clicks hit a hot cache.
3. Independent client fetches for grid / summary / competitors / lockers /
   recommendations, so the map paints progressively instead of waiting for the
   slowest endpoint.

First cold request per province still takes a few seconds.
After that, everything is sub-second.

Deployed solution: [inpost-network-control.vercel.app](https://inpost-network-control.vercel.app/)

If applicable, include:
- a link to the deployed solution
- screenshots of the UI or key outputs
- a short screen recording or demo video

## Technologies

[List the technologies, frameworks, and libraries you used. You can also explain why you decided to use them.]
For the frontend side, I used React. For the server side I used Golang. These are my most used languages. I used PostgreSQL for db and Redis for the cache. As for the libraries and frameworks, I used go-chi and wrote some useful pkg that I needed. Leaflet takes care of the map rendering.

## How to run

### Prerequisites

Anything modern works. Tested on macOS 24 with the following:

- **Go 1.24+** — backend
- **Node 18+** (with `npm`) — frontend
- **Docker** with `docker-compose` — runs Postgres 16 and Redis 7 locally so you don't have to install them yourself
- **make** — orchestrates everything below

That's it. The Makefile copies `.env.example` → `.env` on first run, so no manual config needed for local dev.

### Build & run

```bash
# 1. Clone
git clone git@github.com:SH1NTSU/Inpost-Coverage.git
cd Inpost-Coverage

# 2. Just run this command to start the whole stack
make start
```

## What I would do with more time

[If you had another week, what would you add, refactor, or change? Prioritize — what would you tackle first and why?] Two things: first, continue the current idea, optimize it and add worthiness of placing a point in the desired area. And second, I would go back to my first idea of creating a point predictability engine. I dropped this cause in a week's time I couldn't gather sufficient data for the showcase and functionality of the engine.

## AI usage

[Did you use AI tools (ChatGPT, Copilot, Claude, etc.) while working on this? If yes, describe how — which parts did they help with, and how did you verify and adapt their output?]

I used Claude Code to help me with filling some gaps, fixing bugs and planning with me the implementation. And it helped me with creating the Frontend side. I checked its output myself to ensure that it doesn't do anything stupid.

## Anything else?

[Is there something we should know that doesn't fit the sections above? A design choice that needs context, a creative twist, a rabbit hole you went down — this is your space.] I didn't know if I could but I used a different API for the help but always ensured that the API provided was primarily used. And the toughest thing in this project was the amount of data that needed to be rendered and processed on the frontend.
