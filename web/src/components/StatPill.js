import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
export function StatPill({ label, value, hint, tone = 'neutral' }) {
    return (_jsxs("div", { className: `stat-pill stat-pill--${tone}`, children: [_jsx("div", { className: "stat-pill__label", children: label }), _jsx("div", { className: "stat-pill__value", children: value }), hint && _jsx("div", { className: "stat-pill__hint", children: hint })] }));
}
