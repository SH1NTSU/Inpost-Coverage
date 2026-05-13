import type { LockerStatus } from '../api/contracts';
import { STATUS_COLOR, STATUS_LABEL } from '../lib/status';

export function StatusBadge({ status }: { status: LockerStatus }) {
  return (
    <span
      className="status-badge"
      style={{ backgroundColor: STATUS_COLOR[status] + '22', color: STATUS_COLOR[status] }}
    >
      <span className="status-dot" style={{ backgroundColor: STATUS_COLOR[status] }} />
      {STATUS_LABEL[status]}
    </span>
  );
}
