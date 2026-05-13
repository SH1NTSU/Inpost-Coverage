import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useDeferredValue, useEffect, useMemo } from 'react';
import { CircleMarker, MapContainer, Popup, Rectangle, TileLayer, useMap, } from 'react-leaflet';
import { Link } from 'react-router-dom';
import { TIER_COLOR, networkLabel, formatMeters } from '../../lib/tier';
function MapController({ focus, }) {
    const map = useMap();
    useEffect(() => {
        if (!focus)
            return;
        map.flyTo([focus.lat, focus.lng], 16, { animate: true, duration: 0.7 });
    }, [focus, map]);
    return null;
}
const DEFAULT_CENTER = [52.23, 21.0];
const M_PER_DEG_LAT = 111_320;
const M_PER_DEG_LNG_50 = 71_500;
export function CoverageMap({ cells, cellMeters, inpostLockers, competitors, suggestions, showInpost, showCompetitors, showSuggestions, visibleTiers, focus, }) {
    const deferredCells = useDeferredValue(cells);
    const deferredInpostLockers = useDeferredValue(inpostLockers);
    const deferredCompetitors = useDeferredValue(competitors);
    const deferredSuggestions = useDeferredValue(suggestions);
    const halfLat = cellMeters / 2 / M_PER_DEG_LAT;
    const halfLng = cellMeters / 2 / M_PER_DEG_LNG_50;
    const rectangles = useMemo(() => deferredCells
        .filter((c) => visibleTiers.has(c.tier))
        .map((c) => ({
        key: `${c.lat.toFixed(5)},${c.lng.toFixed(5)}`,
        bounds: [
            [c.lat - halfLat, c.lng - halfLng],
            [c.lat + halfLat, c.lng + halfLng],
        ],
        tier: c.tier,
        inpostM: c.nearest_inpost_m,
        competitorM: c.nearest_competitor_m,
        net: c.nearest_competitor_network,
    })), [deferredCells, halfLat, halfLng, visibleTiers]);
    return (_jsxs(MapContainer, { center: DEFAULT_CENTER, zoom: 12, className: "map", preferCanvas: true, children: [_jsx(MapController, { focus: focus }), _jsx(TileLayer, { attribution: '\u00A9 <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors \u00A9 <a href="https://carto.com/attributions">CARTO</a>', url: "https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png" }), rectangles.map((r) => (_jsx(Rectangle, { bounds: r.bounds, pathOptions: {
                    color: TIER_COLOR[r.tier],
                    fillColor: TIER_COLOR[r.tier],
                    fillOpacity: 0.35,
                    weight: 0,
                }, children: _jsx(Popup, { children: _jsxs("div", { className: "popup popup--coverage", children: [_jsx("strong", { children: r.tier.replace('_', ' ') }), _jsxs("div", { children: ["InPost: ", formatMeters(r.inpostM)] }), _jsxs("div", { children: ["Competitor: ", formatMeters(r.competitorM), " (", networkLabel(r.net), ")"] })] }) }) }, r.key))), showInpost &&
                deferredInpostLockers.map((l) => (_jsx(CircleMarker, { center: [l.latitude, l.longitude], radius: 3, pathOptions: {
                        color: '#3b82f6',
                        fillColor: '#3b82f6',
                        fillOpacity: 0.9,
                        weight: 0,
                    }, children: _jsx(Popup, { children: _jsxs("div", { className: "popup", children: [_jsx("strong", { children: l.inpost_id }), _jsxs("div", { children: [l.street, " ", l.building_no] }), _jsx("div", { className: "popup__city", children: l.city }), _jsx(Link, { to: `/lockers/${l.id}`, className: "popup__link", children: "Details \u2192" })] }) }) }, `i-${l.id}`))), showCompetitors &&
                deferredCompetitors.map((c) => (_jsx(CircleMarker, { center: [c.latitude, c.longitude], radius: 3, pathOptions: {
                        color: '#9ca3af',
                        fillColor: '#9ca3af',
                        fillOpacity: 0.8,
                        weight: 0,
                    }, children: _jsxs(Popup, { children: [_jsx("strong", { children: networkLabel(c.network) }), c.name && _jsx("div", { children: c.name })] }) }, `c-${c.id}`))), showSuggestions &&
                deferredSuggestions.map((s, i) => (_jsx(CircleMarker, { center: [s.lat, s.lng], radius: 10, pathOptions: {
                        color: '#fbbf24',
                        fillColor: '#fbbf24',
                        fillOpacity: 0.9,
                        weight: 2,
                    }, children: _jsx(Popup, { children: _jsxs("div", { className: "popup popup--coverage", children: [_jsxs("strong", { children: ["#", i + 1, " ", s.anchor.brand || s.anchor.name || s.anchor.poi_type] }), _jsx("div", { className: "muted", children: s.anchor.poi_type }), _jsx("div", { className: "popup__city", children: s.reason })] }) }) }, `s-${i}`)))] }));
}
