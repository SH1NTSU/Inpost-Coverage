import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { NavLink } from 'react-router-dom';
import { api } from '../api/client';
export function Layout({ children }) {
    return (_jsxs("div", { className: "app", children: [_jsxs("header", { className: "topbar", children: [_jsxs("div", { className: "topbar__brand", children: [_jsx("span", { className: "logo-dot" }), " InPost Network Console"] }), _jsx("nav", { className: "topbar__nav", children: _jsx(NavLink, { to: "/", end: true, children: "Dashboard" }) }), api.isMock && _jsx("div", { className: "topbar__mock", children: "MOCK DATA" })] }), _jsx("main", { className: "content", children: children })] }));
}
