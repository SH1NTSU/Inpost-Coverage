import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { StatPill } from '../../components/StatPill';
export function SummaryRibbon({ summary }) {
    if (!summary) {
        return (_jsxs("section", { className: "stats-grid", children: [_jsx(StatPill, { label: "Underserved area", value: "\u2026", tone: "bad" }), _jsx(StatPill, { label: "Greenfield gaps", value: "\u2026", tone: "bad" }), _jsx(StatPill, { label: "Competitive gaps", value: "\u2026" }), _jsx(StatPill, { label: "InPost lockers", value: "\u2026", tone: "good" })] }));
    }
    return (_jsxs("section", { className: "stats-grid", children: [_jsx(StatPill, { label: "Underserved area", value: `${summary.underserved_km2.toFixed(1)} km²`, tone: summary.underserved_km2 > 0 ? 'bad' : 'neutral', hint: `${summary.cell_meters}m grid` }), _jsx(StatPill, { label: "Greenfield cells", value: summary.greenfield_cells.toLocaleString(), tone: "bad", hint: "no operator nearby" }), _jsx(StatPill, { label: "Competitive cells", value: summary.competitive_cells.toLocaleString(), hint: "only competitor present" }), _jsx(StatPill, { label: "InPost lockers", value: summary.inpost_lockers.toLocaleString(), tone: "good", hint: `vs ${summary.competitor_lockers.toLocaleString()} competitor` })] }));
}
