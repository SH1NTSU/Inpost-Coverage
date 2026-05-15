import type { CoverageTier } from '../../api/contracts';
import { TIER_COLOR, TIER_LABEL, TIER_DESCRIPTION } from '../../lib/tier';

interface Props {
  visibleTiers: Set<CoverageTier>;
  setVisibleTiers: (s: Set<CoverageTier>) => void;
  showInpost: boolean;
  setShowInpost: (v: boolean) => void;
  showCompetitors: boolean;
  setShowCompetitors: (v: boolean) => void;
  showSuggestions: boolean;
  setShowSuggestions: (v: boolean) => void;
  cellM: number;
  setCellM: (v: number) => void;
}

const TIERS: CoverageTier[] = ['greenfield', 'competitive', 'inpost_only', 'saturated'];

const CELL_SIZES: { value: number; label: string }[] = [
  { value: 800, label: '800 m (fine)' },
  { value: 1500, label: '1.5 km (balanced)' },
  { value: 2500, label: '2.5 km (overview)' },
];

export function LayerControls(props: Props) {
  const toggleTier = (t: CoverageTier) => {
    const next = new Set(props.visibleTiers);
    if (next.has(t)) next.delete(t);
    else next.add(t);
    props.setVisibleTiers(next);
  };

  return (
    <div className="layers">
      <div className="layers__group">
        <span className="layers__title">Cells</span>
        {TIERS.map((t) => (
          <label
            key={t}
            className="layers__chip"
            title={TIER_DESCRIPTION[t]}
            style={{
              borderColor: props.visibleTiers.has(t) ? TIER_COLOR[t] : 'transparent',
              opacity: props.visibleTiers.has(t) ? 1 : 0.5,
            }}
          >
            <input
              type="checkbox"
              checked={props.visibleTiers.has(t)}
              onChange={() => toggleTier(t)}
            />
            <span
              className="layers__swatch"
              style={{ background: TIER_COLOR[t] }}
            />
            {TIER_LABEL[t]}
          </label>
        ))}
      </div>
      <div className="layers__group">
        <span className="layers__title">Grid</span>
        <select
          className="layers__select"
          value={props.cellM}
          onChange={(e) => props.setCellM(Number(e.target.value))}
        >
          {CELL_SIZES.map((s) => (
            <option key={s.value} value={s.value}>
              {s.label}
            </option>
          ))}
        </select>
      </div>
      <div className="layers__group layers__group--points">
        <span className="layers__title">Points</span>
        <label className="layers__chip layers__chip--point">
          <input
            type="checkbox"
            checked={props.showInpost}
            onChange={(e) => props.setShowInpost(e.target.checked)}
          />
          <span className="layers__swatch layers__swatch--inpost" />
          InPost
        </label>
        <label className="layers__chip layers__chip--point">
          <input
            type="checkbox"
            checked={props.showCompetitors}
            onChange={(e) => props.setShowCompetitors(e.target.checked)}
          />
          <span className="layers__swatch layers__swatch--competitors" />
          Competitors
        </label>
        <label className="layers__chip layers__chip--point">
          <input
            type="checkbox"
            checked={props.showSuggestions}
            onChange={(e) => props.setShowSuggestions(e.target.checked)}
          />
          <span className="layers__swatch layers__swatch--suggestions" />
          Suggestions
        </label>
      </div>
    </div>
  );
}
