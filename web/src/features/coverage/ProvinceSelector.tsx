import type { ProvinceInfo } from '../../api/contracts';
import { useProvinces } from '../../api/queries';

interface Props {
  value: string;
  onChange: (province: ProvinceInfo) => void;
}

export function ProvinceSelector({ value, onChange }: Props) {
  const { data: provinces, isLoading } = useProvinces();

  return (
    <div className="province-selector">
      <label htmlFor="province" className="province-selector__label">
        Województwo
      </label>
      <select
        id="province"
        className="province-selector__select"
        value={value}
        disabled={isLoading || !provinces?.length}
        onChange={(e) => {
          const p = provinces?.find((x) => x.name === e.target.value);
          if (p) onChange(p);
        }}
      >
        {isLoading && <option>Loading…</option>}
        {(provinces ?? []).map((p) => (
          <option key={p.name} value={p.name}>
            {p.name} — {p.point_count} lockers
          </option>
        ))}
      </select>
    </div>
  );
}
