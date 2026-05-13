import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useLocker } from '../../api/queries';
import { StatusBadge } from '../../components/StatusBadge';
export default function LockerDetail() {
    const { id } = useParams();
    const lockerId = id ? Number(id) : undefined;
    const { data: locker, isLoading, error } = useLocker(lockerId);
    const [imgFailed, setImgFailed] = useState(false);
    if (isLoading)
        return _jsx("p", { className: "content-pad", children: "Loading\u2026" });
    if (error || !locker) {
        return (_jsxs("div", { className: "content-pad", children: [_jsx("p", { children: "Locker not found." }), _jsx(Link, { to: "/", children: "\u2190 back to dashboard" })] }));
    }
    const showImage = !!locker.image_url && !imgFailed;
    return (_jsxs("article", { className: "locker-detail", children: [_jsx(Link, { to: "/", className: "back-link", children: "\u2190 back" }), _jsxs("header", { className: "locker-detail__head", children: [_jsxs("div", { className: "locker-detail__title", children: [_jsx("h1", { children: locker.inpost_id }), _jsxs("p", { className: "muted", children: [locker.street, " ", locker.building_no, ", ", locker.post_code, " ", locker.city] })] }), _jsx("div", { className: "locker-detail__head-actions", children: _jsx(StatusBadge, { status: locker.current_status }) })] }), showImage && (_jsx("figure", { className: "locker-detail__photo", children: _jsx("img", { src: locker.image_url, alt: `Paczkomat ${locker.inpost_id}`, loading: "lazy", onError: () => setImgFailed(true) }) })), _jsxs("section", { className: "locker-detail__meta", children: [_jsxs("div", { children: [_jsx("span", { className: "label", children: "24/7" }), " ", locker.location_247 ? 'yes' : 'no'] }), _jsxs("div", { children: [_jsx("span", { className: "label", children: "New generation" }), " ", locker.is_next ? 'yes' : 'no'] }), _jsxs("div", { children: [_jsx("span", { className: "label", children: "Type" }), " ", locker.location_type] }), _jsxs("div", { children: [_jsx("span", { className: "label", children: "Physical" }), " ", locker.physical_type] })] })] }));
}
