import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useCallback, useEffect, useState } from 'react';
import { useCompetitors, useCoverageGridCells, useCoverageRecommendations, useCoverageSummary, useLockers, useProvinces, } from '../../api/queries';
import { CoverageMap } from './CoverageMap';
import { LayerControls } from './LayerControls';
import { ProvinceSelector } from './ProvinceSelector';
import { RecommendationsPanel } from './RecommendationsPanel';
import { SummaryRibbon } from './SummaryRibbon';
const DEFAULT_TIERS = ['greenfield', 'competitive'];
export default function CoverageView() {
    const [cellM] = useState(800);
    const [visibleTiers, setVisibleTiers] = useState(new Set(DEFAULT_TIERS));
    const [showInpost, setShowInpost] = useState(true);
    const [showCompetitors, setShowCompetitors] = useState(true);
    const [showSuggestions, setShowSuggestions] = useState(true);
    const [focus, setFocus] = useState(null);
    const [province, setProvince] = useState(null);
    const { data: provinces } = useProvinces();
    useEffect(() => {
        if (!province && provinces && provinces.length > 0) {
            const next = provinces[0];
            setProvince(next);
            setFocus({
                lat: next.center_lat,
                lng: next.center_lng,
                tick: Date.now(),
            });
        }
    }, [province, provinces]);
    const provinceName = province?.name ?? '';
    const summary = useCoverageSummary(provinceName, cellM);
    const gridCells = useCoverageGridCells(provinceName, cellM, !!summary.data);
    const recs = useCoverageRecommendations(provinceName, 5);
    const competitors = useCompetitors(provinceName);
    const lockers = useLockers(undefined, provinceName);
    const focusOn = useCallback((s) => {
        setFocus({ lat: s.lat, lng: s.lng, tick: Date.now() });
    }, []);
    const handleProvinceChange = useCallback((p) => {
        setProvince(p);
        setFocus({ lat: p.center_lat, lng: p.center_lng, tick: Date.now() });
    }, []);
    return (_jsxs("section", { className: "coverage", children: [_jsx("header", { className: "coverage__head", children: _jsxs("div", { className: "coverage__head-row", children: [_jsxs("div", { children: [_jsx("h1", { children: "Coverage gaps & upgrade map" }), _jsx("p", { className: "muted", children: "Where InPost's network is thin, who else is there, and which existing lockers are due for replacement." })] }), _jsx(ProvinceSelector, { value: provinceName, onChange: handleProvinceChange })] }) }), _jsx(SummaryRibbon, { summary: summary.data }), _jsx(LayerControls, { visibleTiers: visibleTiers, setVisibleTiers: setVisibleTiers, showInpost: showInpost, setShowInpost: setShowInpost, showCompetitors: showCompetitors, setShowCompetitors: setShowCompetitors, showSuggestions: showSuggestions, setShowSuggestions: setShowSuggestions }), _jsxs("div", { className: "coverage__body", children: [_jsxs("div", { className: "coverage__map", children: [(summary.isLoading || gridCells.isLoading) && (_jsx("div", { className: "map-loader", children: "loading coverage\u2026" })), _jsx(CoverageMap, { cells: gridCells.data ?? [], cellMeters: summary.data?.cell_meters ?? cellM, inpostLockers: lockers.data ?? [], competitors: competitors.data ?? [], suggestions: recs.data?.new_points ?? [], showInpost: showInpost, showCompetitors: showCompetitors, showSuggestions: showSuggestions, visibleTiers: visibleTiers, focus: focus })] }), _jsx(RecommendationsPanel, { data: recs.data, isLoading: recs.isLoading, onPickSuggestion: focusOn })] }), _jsxs("footer", { className: "coverage__footer", children: ["Coverage analysis based on InPost Points API + community-maintained", ' ', _jsx("a", { href: "https://www.openstreetmap.org/copyright", target: "_blank", rel: "noreferrer", children: "OpenStreetMap" }), ' ', "competitor data. Counts are a lower bound \u2014 some competitor lockers may not yet be tagged in OSM."] })] }));
}
