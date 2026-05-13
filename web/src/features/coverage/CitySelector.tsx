import type { CityInfo } from '../../api/contracts';
import { useCities } from '../../api/queries';

interface Props {
  value: string;
  onChange: (city: CityInfo) => void;
}

export function CitySelector({ value, onChange }: Props) {
  const { data: cities, isLoading } = useCities();

  return (
    <div className="city-selector">
      <label htmlFor="city" className="city-selector__label">
        City
      </label>
      <select
        id="city"
        className="city-selector__select"
        value={value}
        disabled={isLoading || !cities?.length}
        onChange={(e) => {
          const c = cities?.find((x) => x.name === e.target.value);
          if (c) onChange(c);
        }}
      >
        {isLoading && <option>Loading…</option>}
        {(cities ?? []).map((c) => (
          <option key={c.name} value={c.name}>
            {c.name} — {c.point_count} lockers
          </option>
        ))}
      </select>
    </div>
  );
}
