import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { STATUS_COLOR, STATUS_LABEL } from '../lib/status';
export function StatusBadge({ status }) {
    return (_jsxs("span", { className: "status-badge", style: { backgroundColor: STATUS_COLOR[status] + '22', color: STATUS_COLOR[status] }, children: [_jsx("span", { className: "status-dot", style: { backgroundColor: STATUS_COLOR[status] } }), STATUS_LABEL[status]] }));
}
