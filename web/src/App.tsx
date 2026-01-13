import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import axios from 'axios'
import { AlertCircle, CheckCircle, ExternalLink, RefreshCw, Anchor, Package, Info, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react'
import type { ComparisonResult, DriftStatus } from './types'

// Setup Axios (Assume backend is on 8080)
// In dev, we might need a proxy or direct URL.
const API_URL = import.meta.env.DEV ? 'http://localhost:8080/api/v1' : '/api/v1'

const fetchReleases = async (): Promise<ComparisonResult[]> => {
  const { data } = await axios.get(`${API_URL}/releases`)
  return data
}

function StatusBadge({ status }: { status: DriftStatus }) {
  switch (status) {
    case 'SYNC':
      return (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
          <CheckCircle className="w-3 h-3 mr-1" /> Sync
        </span>
      )
    case 'PATCH_DRIFT':
      return (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
          <Info className="w-3 h-3 mr-1" /> Patch Available
        </span>
      )
    case 'MINOR_DRIFT':
      return (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
          <AlertCircle className="w-3 h-3 mr-1" /> Minor Drift
        </span>
      )
    case 'MAJOR_DRIFT':
      return (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
          <AlertCircle className="w-3 h-3 mr-1" /> Major Drift
        </span>
      )
    default:
      return (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
          <AlertCircle className="w-3 h-3 mr-1" /> Unknown
        </span>
      )
  }
}

// Sort helper
type SortKey = 'status' | 'release' | 'namespace';
type SortDirection = 'asc' | 'desc';

function App() {
  const { data: releases, isLoading, isError, refetch, isRefetching } = useQuery({
    queryKey: ['releases'],
    queryFn: fetchReleases,
    refetchInterval: 60000, // Auto-refresh every 1 min
  })

  // Sorting State
  const [sortConfig, setSortConfig] = useState<{ key: SortKey; direction: SortDirection } | null>({ key: 'status', direction: 'asc' });

  const handleSort = (key: SortKey) => {
    let direction: SortDirection = 'asc';
    if (sortConfig && sortConfig.key === key && sortConfig.direction === 'asc') {
      direction = 'desc';
    }
    setSortConfig({ key, direction });
  };

  // Status Priority Map (Lower number = Higher Severity for default ASC sort to show problems first)
  const statusPriority: Record<string, number> = {
    'MAJOR_DRIFT': 1,
    'MINOR_DRIFT': 2,
    'PATCH_DRIFT': 3,
    'SYNC': 4,
    'UNKNOWN': 5,
  };

  const sortedReleases = [...(releases || [])].sort((a, b) => {
    if (!sortConfig) return 0;

    let comparison = 0;
    if (sortConfig.key === 'status') {
      const pA = statusPriority[a.status] || 99;
      const pB = statusPriority[b.status] || 99;
      comparison = pA - pB;
    } else if (sortConfig.key === 'release') {
      comparison = a.release.name.localeCompare(b.release.name);
    } else if (sortConfig.key === 'namespace') {
      comparison = a.release.namespace.localeCompare(b.release.namespace);
    }

    return sortConfig.direction === 'asc' ? comparison : -comparison;
  });

  const getSortIcon = (key: SortKey) => {
    if (sortConfig?.key !== key) return <ArrowUpDown className="w-4 h-4 ml-1 text-slate-400 opacity-50 group-hover:opacity-100 transition-opacity" />;
    return sortConfig.direction === 'asc'
      ? <ArrowUp className="w-4 h-4 ml-1 text-blue-600" />
      : <ArrowDown className="w-4 h-4 ml-1 text-blue-600" />;
  };

  return (
    <div className="min-h-screen bg-slate-50 font-sans text-slate-900">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Anchor className="w-6 h-6 text-blue-600" />
            <h1 className="text-xl font-bold bg-gradient-to-r from-blue-600 to-cyan-500 bg-clip-text text-transparent">
              KubeScout
            </h1>
          </div>
          <div className="flex items-center gap-4">
            <button
              onClick={() => refetch()}
              className="p-2 text-slate-500 hover:text-blue-600 hover:bg-slate-100 rounded-full transition-colors"
              title="Refresh"
            >
              <RefreshCw className={`w-5 h-5 ${isRefetching ? 'animate-spin' : ''}`} />
            </button>
            <a
              href="https://github.com/haedalwang/kubescout"
              target="_blank"
              rel="noreferrer"
              className="text-slate-500 hover:text-slate-900"
            >
              <svg role="img" viewBox="0 0 24 24" fill="currentColor" className="w-6 h-6">
                <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
              </svg>
            </a>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">

        {/* Status Overview */}
        <div className="mb-8">
          <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6 flex flex-col md:flex-row items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-slate-900 flex items-center gap-2">
                <Package className="w-5 h-5 text-slate-500" />
                Installed Charts
              </h2>
              <p className="text-sm text-slate-500 mt-1">Overview of your Helm releases and drift status.</p>
            </div>
            <div className="mt-4 md:mt-0 flex gap-4">
              <div className="text-right">
                <span className="text-2xl font-bold text-slate-900 block">{releases?.length || 0}</span>
                <span className="text-xs text-slate-500 uppercase font-medium">Total</span>
              </div>
            </div>
          </div>
        </div>

        {/* Error State */}
        {isError && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6 flex items-center gap-2">
            <AlertCircle className="w-5 h-5" />
            <span>Failed to load releases. Is the backend running?</span>
          </div>
        )}

        {/* Loading State */}
        {isLoading && (
          <div className="p-12 text-center text-slate-500">
            Loading releases...
          </div>
        )}

        {/* Table */}
        {!isLoading && !isError && (
          <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-slate-200">
                <thead className="bg-slate-50 text-slate-500 text-xs font-semibold uppercase tracking-wider">
                  <tr>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left cursor-pointer hover:bg-slate-100 transition-colors group"
                      onClick={() => handleSort('status')}
                    >
                      <div className="flex items-center">
                        Status
                        {getSortIcon('status')}
                      </div>
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left cursor-pointer hover:bg-slate-100 transition-colors group"
                      onClick={() => handleSort('release')}
                    >
                      <div className="flex items-center">
                        Release
                        {getSortIcon('release')}
                      </div>
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left cursor-pointer hover:bg-slate-100 transition-colors group"
                      onClick={() => handleSort('namespace')}
                    >
                      <div className="flex items-center">
                        Namespace
                        {getSortIcon('namespace')}
                      </div>
                    </th>
                    <th scope="col" className="px-6 py-3 text-left">Current Version</th>
                    <th scope="col" className="px-6 py-3 text-left">Latest Version</th>
                    <th scope="col" className="px-6 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-slate-200">
                  {sortedReleases.map((item: ComparisonResult) => (
                    <tr key={`${item.release.namespace}/${item.release.name}`} className="hover:bg-slate-50 transition-colors">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <StatusBadge status={item.status} />
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          {item.release.icon && (
                            <img className="h-8 w-8 rounded-md mr-3 object-contain bg-slate-100" src={item.release.icon} alt="" />
                          )}
                          <div>
                            <div className="text-sm font-medium text-slate-900">{item.release.name}</div>
                            <div className="text-xs text-slate-500">{item.release.chart_name}</div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
                        {item.release.namespace}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-slate-900 font-mono">{item.release.chart_version}</div>
                        <div className="text-xs text-slate-400">App: {item.release.app_version}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-slate-900 font-mono">{item.latest_version}</div>
                        <div className="text-xs text-slate-400">App: {item.latest_app_version}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <a href={item.upstream_url} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-900 inline-flex items-center gap-1">
                          Upstream <ExternalLink className="w-3 h-3" />
                        </a>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}

export default App
