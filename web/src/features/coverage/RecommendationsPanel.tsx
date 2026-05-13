import { Link } from 'react-router-dom';

import type {
  CoverageRecommendations,
  GapSuggestion,
  UpgradeCandidate,
} from '../../api/contracts';
import { TIER_COLOR, TIER_LABEL, formatMeters, networkLabel } from '../../lib/tier';

interface Props {
  data: CoverageRecommendations | undefined;
  isLoading: boolean;
  onPickSuggestion?: (s: GapSuggestion) => void;
}

function anchorLabel(a: { brand: string; name: string; poi_type: string }) {
  return a.brand || a.name || a.poi_type;
}

export function RecommendationsPanel({ data, isLoading, onPickSuggestion }: Props) {
  return (
    <aside className="recs">
      <section className="recs__group">
        <h2>🆕 Suggested anchor locations</h2>
        <p className="recs__subtitle muted">
          Real places far from any InPost locker — attach a Paczkomat here.
        </p>
        {isLoading && <p className="muted">Calculating…</p>}
        {!isLoading && (!data || data.new_points.length === 0) && (
          <p className="muted">No isolated anchor POIs in this view.</p>
        )}
        <ol className="recs__list">
          {(data?.new_points ?? []).map((s, i) => (
            <li key={i} className="rec-card">
              <button
                type="button"
                className="rec-card__head"
                onClick={() => onPickSuggestion?.(s)}
                aria-label={`Focus map on suggestion ${i + 1}`}
              >
                <span
                  className="rec-card__pill"
                  style={{ background: TIER_COLOR[s.tier] }}
                >
                  #{i + 1} {TIER_LABEL[s.tier]}
                </span>
                <span className="rec-card__anchor-headline">
                  <span className="anchor-type">{s.anchor.poi_type}</span>
                  <strong>{anchorLabel(s.anchor)}</strong>
                </span>
              </button>
              <p className="rec-card__reason">{s.reason}</p>
              <dl className="rec-card__metrics">
                <div>
                  <dt>Nearest InPost</dt>
                  <dd>{formatMeters(s.nearest_inpost_m)}</dd>
                </div>
                <div>
                  <dt>Nearest competitor</dt>
                  <dd>
                    {formatMeters(s.nearest_competitor_m)}{' '}
                    <span className="muted">({networkLabel(s.nearest_competitor_network)})</span>
                  </dd>
                </div>
              </dl>
              {s.nearby_anchors.length > 0 && (
                <div className="rec-card__anchors">
                  <span className="rec-card__anchors-title">Other anchors within 250 m</span>
                  <ul>
                    {s.nearby_anchors.slice(0, 4).map((a, k) => (
                      <li key={k}>
                        <span className="anchor-type">{a.poi_type}</span>
                        <span className="anchor-name">{anchorLabel(a)}</span>
                        <span className="anchor-dist">{formatMeters(a.distance_m)}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </li>
          ))}
        </ol>
      </section>

      <section className="recs__group">
        <h2>🔁 Lockers worth upgrading</h2>
        <p className="recs__subtitle muted">
          Old-generation units, ranked by surrounding competitor pressure.
        </p>
        {isLoading && <p className="muted">Calculating…</p>}
        {!isLoading && (!data || data.upgrades.length === 0) && (
          <p className="muted">No old-generation lockers in this view.</p>
        )}
        <ol className="recs__list">
          {(data?.upgrades ?? []).map((u: UpgradeCandidate, i: number) => (
            <li key={u.locker.id} className="rec-card">
              <div className="rec-card__head rec-card__head--row">
                <Link to={`/lockers/${u.locker.id}`} className="rec-card__id">
                  #{i + 1} {u.locker.inpost_id}
                </Link>
                <span className="rec-card__score">
                  score {(u.score * 100).toFixed(0)}
                </span>
              </div>
              <p className="muted">{u.locker.street}, {u.locker.city}</p>
              <dl className="rec-card__metrics">
                <div>
                  <dt>Generation</dt>
                  <dd>{u.is_next ? 'new' : 'old'}</dd>
                </div>
                <div>
                  <dt>Competitors ≤ 400m</dt>
                  <dd>{u.competitor_pressure}</dd>
                </div>
                <div>
                  <dt>24/7</dt>
                  <dd>{u.locker.location_247 ? 'yes' : 'no'}</dd>
                </div>
              </dl>
              {u.reasons.length > 0 && (
                <ul className="rec-card__reasons">
                  {u.reasons.map((r) => (
                    <li key={r}>{r}</li>
                  ))}
                </ul>
              )}
            </li>
          ))}
        </ol>
      </section>
    </aside>
  );
}
