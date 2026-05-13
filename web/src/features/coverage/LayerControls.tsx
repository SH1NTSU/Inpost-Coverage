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
}

const TIERS: CoverageTier[] = ['greenfield', 'competitive', 'inpost_only', 'saturated'];

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
        <span className="layers__title">Points</span>
        <label className="layers__chip">
          <input
            type="checkbox"
            checked={props.showInpost}
            onChange={(e) => props.setShowInpost(e.target.checked)}
          />
          <span className="layers__swatch" style={{ background: '#3b82f6' }} />
          InPost
        </label>
        <label className="layers__chip">
          <input
            type="checkbox"
            checked={props.showCompetitors}
            onChange={(e) => props.setShowCompetitors(e.target.checked)}
          />
          <span className="layers__swatch" style={{ background: '#9ca3af' }} />
          Competitors
        </label>
        <label className="layers__chip">
          <input
            type="checkbox"
            checked={props.showSuggestions}
            onChange={(e) => props.setShowSuggestions(e.target.checked)}
          />
          <span className="layers__swatch" style={{ background: '#fbbf24' }} />
          Suggestions
        </label>
      </div>
    </div>
  );
}
