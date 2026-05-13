import { Route, Routes } from 'react-router-dom';

import { Layout } from './components/Layout';
import CoverageView from './features/coverage/CoverageView';
import LockerDetail from './features/locker/LockerDetail';

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<CoverageView />} />
        <Route path="/lockers/:id" element={<LockerDetail />} />
        <Route path="*" element={<p className="content-pad">Page not found.</p>} />
      </Routes>
    </Layout>
  );
}
