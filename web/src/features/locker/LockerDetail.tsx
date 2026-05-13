import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';

import { useLocker } from '../../api/queries';
import { StatusBadge } from '../../components/StatusBadge';

export default function LockerDetail() {
  const { id } = useParams<{ id: string }>();
  const lockerId = id ? Number(id) : undefined;
  const { data: locker, isLoading, error } = useLocker(lockerId);
  const [imgFailed, setImgFailed] = useState(false);

  if (isLoading) return <p className="content-pad">Loading…</p>;
  if (error || !locker) {
    return (
      <div className="content-pad">
        <p>Locker not found.</p>
        <Link to="/">← back to dashboard</Link>
      </div>
    );
  }

  const showImage = !!locker.image_url && !imgFailed;

  return (
    <article className="locker-detail">
      <Link to="/" className="back-link">← back</Link>

      <header className="locker-detail__head">
        <div className="locker-detail__title">
          <h1>{locker.inpost_id}</h1>
          <p className="muted">
            {locker.street} {locker.building_no}, {locker.post_code} {locker.city}
          </p>
        </div>
        <div className="locker-detail__head-actions">
          <StatusBadge status={locker.current_status} />
        </div>
      </header>

      {showImage && (
        <figure className="locker-detail__photo">
          <img
            src={locker.image_url}
            alt={`Paczkomat ${locker.inpost_id}`}
            loading="lazy"
            onError={() => setImgFailed(true)}
          />
        </figure>
      )}

      <section className="locker-detail__meta">
        <div><span className="label">24/7</span> {locker.location_247 ? 'yes' : 'no'}</div>
        <div><span className="label">New generation</span> {locker.is_next ? 'yes' : 'no'}</div>
        <div><span className="label">Type</span> {locker.location_type}</div>
        <div><span className="label">Physical</span> {locker.physical_type}</div>
      </section>
    </article>
  );
}
