export function formatPercent(v, digits = 1) {
    return `${(v * 100).toFixed(digits)}%`;
}
export function formatDuration(seconds) {
    if (seconds < 60)
        return `${seconds}s`;
    if (seconds < 3600)
        return `${Math.floor(seconds / 60)}m`;
    if (seconds < 86400) {
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        return m === 0 ? `${h}h` : `${h}h ${m}m`;
    }
    const d = Math.floor(seconds / 86400);
    const h = Math.floor((seconds % 86400) / 3600);
    return h === 0 ? `${d}d` : `${d}d ${h}h`;
}
export function formatRelative(iso) {
    if (!iso)
        return '—';
    const diff = Date.now() - new Date(iso).getTime();
    if (diff < 60_000)
        return 'just now';
    return `${formatDuration(Math.floor(diff / 1000))} ago`;
}
export function formatDateTime(iso) {
    if (!iso)
        return '—';
    return new Date(iso).toLocaleString();
}
