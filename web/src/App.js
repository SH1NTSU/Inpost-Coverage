import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { Route, Routes } from 'react-router-dom';
import { Layout } from './components/Layout';
import CoverageView from './features/coverage/CoverageView';
import LockerDetail from './features/locker/LockerDetail';
export default function App() {
    return (_jsx(Layout, { children: _jsxs(Routes, { children: [_jsx(Route, { path: "/", element: _jsx(CoverageView, {}) }), _jsx(Route, { path: "/lockers/:id", element: _jsx(LockerDetail, {}) }), _jsx(Route, { path: "*", element: _jsx("p", { className: "content-pad", children: "Page not found." }) })] }) }));
}
