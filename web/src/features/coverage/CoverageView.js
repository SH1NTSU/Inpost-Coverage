import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useCallback, useEffect, useState } from 'react';
import { useCities, useCompetitors, useCoverageGridCells, useCoverageRecommendations, useCoverageSummary, useLockers, } from '../../api/queries';
import { CitySelector } from './CitySelector';
import { CoverageMap } from './CoverageMap';
import { LayerControls } from './LayerControls';
import { RecommendationsPanel } from './RecommendationsPanel';
import { SummaryRibbon } from './SummaryRibbon';
const DEFAULT_TIERS = ['greenfield', 'competitive'];
export default function CoverageView() {
    const [cellM] = useState(400);
    const [visibleTiers, setVisibleTiers] = useState(new Set(DEFAULT_TIERS));
    const [showInpost, setShowInpost] = useState(true);
    const [showCompetitors, setShowCompetitors] = useState(true);
    const [showSuggestions, setShowSuggestions] = useState(true);
    const [focus, setFocus] = useState(null);
    const [city, setCity] = useState(null);
    const { data: cities } = useCities();
    useEffect(() => {
        if (!city && cities && cities.length > 0) {
            const next = cities[0];
            setCity(next);
            setFocus({
                lat: next.center_lat,
                lng: next.center_lng,
                tick: Date.now(),
            });
        }
    }, [city, cities]);
    const cityName = city?.name ?? '';
    const summary = useCoverageSummary(cityName, cellM);
    const gridCells = useCoverageGridCells(cityName, cellM, !!summary.data);
    const recs = useCoverageRecommendations(cityName, 5);
    const competitors = useCompetitors(cityName);
    const lockers = useLockers(undefined, cityName);
    const focusOn = useCallback((s) => {
        setFocus({ lat: s.lat, lng: s.lng, tick: Date.now() });
    }, []);
    const handleCityChange = useCallback((c) => {
        setCity(c);
        setFocus({ lat: c.center_lat, lng: c.center_lng, tick: Date.now() });
    }, []);
    return (_jsxs("section", { className: "coverage", children: [_jsx("header", { className: "coverage__head", children: _jsxs("div", { className: "coverage__head-row", children: [_jsxs("div", { children: [_jsx("h1", { children: "Coverage gaps & upgrade map" }), _jsx("p", { className: "muted", children: "Where InPost's network is thin, who else is there, and which existing lockers are due for replacement." })] }), _jsx(CitySelector, { value: cityName, onChange: handleCityChange })] }) }), _jsx(SummaryRibbon, { summary: summary.data }), _jsx(LayerControls, { visibleTiers: visibleTiers, setVisibleTiers: setVisibleTiers, showInpost: showInpost, setShowInpost: setShowInpost, showCompetitors: showCompetitors, setShowCompetitors: setShowCompetitors, showSuggestions: showSuggestions, setShowSuggestions: setShowSuggestions }), _jsxs("div", { className: "coverage__body", children: [_jsxs("div", { className: "coverage__map", children: [(summary.isLoading || gridCells.isLoading) && (_jsx("div", { className: "map-loader", children: "loading coverage\u2026" })), _jsx(CoverageMap, { cells: gridCells.data ?? [], cellMeters: summary.data?.cell_meters ?? cellM, inpostLockers: lockers.data ?? [], competitors: competitors.data ?? [], suggestions: recs.data?.new_points ?? [], showInpost: showInpost, showCompetitors: showCompetitors, showSuggestions: showSuggestions, visibleTiers: visibleTiers, focus: focus })] }), _jsx(RecommendationsPanel, { data: recs.data, isLoading: recs.isLoading, onPickSuggestion: focusOn })] }), _jsxs("footer", { className: "coverage__footer", children: ["Coverage analysis based on InPost Points API + community-maintained", ' ', _jsx("a", { href: "https://www.openstreetmap.org/copyright", target: "_blank", rel: "noreferrer", children: "OpenStreetMap" }), ' ', "competitor data. Counts are a lower bound \u2014 some competitor lockers may not yet be tagged in OSM."] })] }));
}
