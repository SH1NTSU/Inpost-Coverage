import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useCities } from '../../api/queries';
export function CitySelector({ value, onChange }) {
    const { data: cities, isLoading } = useCities();
    return (_jsxs("div", { className: "city-selector", children: [_jsx("label", { htmlFor: "city", className: "city-selector__label", children: "City" }), _jsxs("select", { id: "city", className: "city-selector__select", value: value, disabled: isLoading || !cities?.length, onChange: (e) => {
                    const c = cities?.find((x) => x.name === e.target.value);
                    if (c)
                        onChange(c);
                }, children: [isLoading && _jsx("option", { children: "Loading\u2026" }), (cities ?? []).map((c) => (_jsxs("option", { value: c.name, children: [c.name, " \u2014 ", c.point_count, " lockers"] }, c.name)))] })] }));
}
