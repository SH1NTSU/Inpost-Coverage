import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { TIER_COLOR, TIER_LABEL, TIER_DESCRIPTION } from '../../lib/tier';
const TIERS = ['greenfield', 'competitive', 'inpost_only', 'saturated'];
export function LayerControls(props) {
    const toggleTier = (t) => {
        const next = new Set(props.visibleTiers);
        if (next.has(t))
            next.delete(t);
        else
            next.add(t);
        props.setVisibleTiers(next);
    };
    return (_jsxs("div", { className: "layers", children: [_jsxs("div", { className: "layers__group", children: [_jsx("span", { className: "layers__title", children: "Cells" }), TIERS.map((t) => (_jsxs("label", { className: "layers__chip", title: TIER_DESCRIPTION[t], style: {
                            borderColor: props.visibleTiers.has(t) ? TIER_COLOR[t] : 'transparent',
                            opacity: props.visibleTiers.has(t) ? 1 : 0.5,
                        }, children: [_jsx("input", { type: "checkbox", checked: props.visibleTiers.has(t), onChange: () => toggleTier(t) }), _jsx("span", { className: "layers__swatch", style: { background: TIER_COLOR[t] } }), TIER_LABEL[t]] }, t)))] }), _jsxs("div", { className: "layers__group", children: [_jsx("span", { className: "layers__title", children: "Points" }), _jsxs("label", { className: "layers__chip", children: [_jsx("input", { type: "checkbox", checked: props.showInpost, onChange: (e) => props.setShowInpost(e.target.checked) }), _jsx("span", { className: "layers__swatch", style: { background: '#3b82f6' } }), "InPost"] }), _jsxs("label", { className: "layers__chip", children: [_jsx("input", { type: "checkbox", checked: props.showCompetitors, onChange: (e) => props.setShowCompetitors(e.target.checked) }), _jsx("span", { className: "layers__swatch", style: { background: '#9ca3af' } }), "Competitors"] }), _jsxs("label", { className: "layers__chip", children: [_jsx("input", { type: "checkbox", checked: props.showSuggestions, onChange: (e) => props.setShowSuggestions(e.target.checked) }), _jsx("span", { className: "layers__swatch", style: { background: '#fbbf24' } }), "Suggestions"] })] })] }));
}
