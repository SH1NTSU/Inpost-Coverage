import type { CoverageSummary } from '../../api/contracts';
import { StatPill } from '../../components/StatPill';

export function SummaryRibbon({ summary }: { summary: CoverageSummary | undefined }) {
  if (!summary) {
    return (
      <section className="stats-grid">
        <StatPill label="Underserved area" value="…" tone="bad" />
        <StatPill label="Greenfield gaps" value="…" tone="bad" />
        <StatPill label="Competitive gaps" value="…" />
        <StatPill label="InPost lockers" value="…" tone="good" />
      </section>
    );
  }
  return (
    <section className="stats-grid">
      <StatPill
        label="Underserved area"
        value={`${summary.underserved_km2.toFixed(1)} km²`}
        tone={summary.underserved_km2 > 0 ? 'bad' : 'neutral'}
        hint={`${summary.cell_meters}m grid`}
      />
      <StatPill
        label="Greenfield cells"
        value={summary.greenfield_cells.toLocaleString()}
        tone="bad"
        hint="no operator mapped nearby"
      />
      <StatPill
        label="Competitive cells"
        value={summary.competitive_cells.toLocaleString()}
        hint="only competitor present"
      />
      <StatPill
        label="InPost lockers"
        value={summary.inpost_lockers.toLocaleString()}
        tone="good"
        hint={`vs ${summary.competitor_lockers.toLocaleString()} competitor`}
      />
    </section>
  );
}
