import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useProvinces } from '../../api/queries';
export function ProvinceSelector({ value, onChange }) {
    const { data: provinces, isLoading } = useProvinces();
    return (_jsxs("div", { className: "province-selector", children: [_jsx("label", { htmlFor: "province", className: "province-selector__label", children: "Wojew\u00F3dztwo" }), _jsxs("select", { id: "province", className: "province-selector__select", value: value, disabled: isLoading || !provinces?.length, onChange: (e) => {
                    const p = provinces?.find((x) => x.name === e.target.value);
                    if (p)
                        onChange(p);
                }, children: [isLoading && _jsx("option", { children: "Loading\u2026" }), (provinces ?? []).map((p) => (_jsxs("option", { value: p.name, children: [p.name, " \u2014 ", p.point_count, " lockers"] }, p.name)))] })] }));
}
