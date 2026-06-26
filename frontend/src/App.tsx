import { Routes, Route, Navigate } from 'react-router-dom';
import { type ReactNode } from 'react';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { OrgProvider } from '@/context/OrgContext';
import { Layout } from '@/components/Layout';
import Login from '@/pages/Login';
import Register from '@/pages/Register';
import Dashboard from '@/pages/Dashboard';
import Organizations from '@/pages/Organizations';
import Projects from '@/pages/Projects';
import AssetInventory from '@/pages/AssetInventory';
import AssetDetails from '@/pages/AssetDetails';
import Reports from '@/pages/Reports';
import Settings from '@/pages/Settings';
import Profile from '@/pages/Profile';

function Shell({ children }: { children: ReactNode }) {
  return (
    <ProtectedRoute>
      <OrgProvider>
        <Layout>{children}</Layout>
      </OrgProvider>
    </ProtectedRoute>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/" element={<Shell><Dashboard /></Shell>} />
      <Route path="/organizations" element={<Shell><Organizations /></Shell>} />
      <Route path="/projects" element={<Shell><Projects /></Shell>} />
      <Route path="/assets" element={<Shell><AssetInventory /></Shell>} />
      <Route path="/assets/:projectId/:assetId" element={<Shell><AssetDetails /></Shell>} />
      <Route path="/reports" element={<Shell><Reports /></Shell>} />
      <Route path="/settings" element={<Shell><Settings /></Shell>} />
      <Route path="/profile" element={<Shell><Profile /></Shell>} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
