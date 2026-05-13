import type { ReactNode } from 'react';
import { NavLink } from 'react-router-dom';

import { api } from '../api/client';

export function Layout({ children }: { children: ReactNode }) {
  return (
    <div className="app">
      <header className="topbar">
        <div className="topbar__brand">
          <span className="logo-dot" /> InPost Network Console
        </div>
        <nav className="topbar__nav">
          <NavLink to="/" end>Dashboard</NavLink>
        </nav>
        {api.isMock && <div className="topbar__mock">MOCK DATA</div>}
      </header>
      <main className="content">{children}</main>
    </div>
  );
}
