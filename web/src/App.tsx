import { useQuery } from '@tanstack/react-query'
import axios from 'axios'
import { AlertCircle, CheckCircle, ExternalLink, RefreshCw, Anchor, Package } from 'lucide-react'
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

function App() {
  const { data: releases, isLoading, isError, refetch, isRefetching } = useQuery({
    queryKey: ['releases'],
    queryFn: fetchReleases,
    refetchInterval: 60000, // Auto-refresh every 1 min
  })

  return (
    <div className="min-h-screen bg-slate-50 font-sans text-slate-900">
      {/* Header */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="bg-blue-600 p-1.5 rounded-lg">
              <Anchor className="w-6 h-6 text-white" />
            </div>
            <h1 className="text-xl font-bold tracking-tight text-slate-900">KubeScout</h1>
          </div>
          <div className="flex items-center space-x-4">
            <button
              onClick={() => refetch()}
              disabled={isRefetching}
              className={`p-2 rounded-full hover:bg-slate-100 transition-colors ${isRefetching ? 'animate-spin' : ''}`}
            >
              <RefreshCw className="w-5 h-5 text-slate-600" />
            </button>
            <a href="https://github.com/haedalwang/kubescout" target="_blank" rel="noopener noreferrer" className="text-slate-500 hover:text-slate-900">
              <svg viewBox="0 0 24 24" className="h-6 w-6 fill-current" aria-hidden="true"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" /></svg>
            </a>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">

        {/* Stats / Overview could go here */}

        <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
          <div className="px-6 py-4 border-b border-slate-200 bg-slate-50 flex justify-between items-center">
            <h2 className="text-base font-semibold text-slate-800 flex items-center gap-2">
              <Package className="w-4 h-4" /> Installed Charts
            </h2>
            <span className="text-sm text-slate-500">
              Total: <span className="font-medium text-slate-900">{releases?.length || 0}</span>
            </span>
          </div>

          {isLoading && (
            <div className="p-12 text-center text-slate-500">
              Loading releases...
            </div>
          )}

          {isError && (
            <div className="p-12 text-center text-red-500">
              Failed to load releases. Is the backend running?
            </div>
          )}

          {!isLoading && !isError && (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-slate-200">
                <thead className="bg-slate-50">
                  <tr>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Status</th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Release / Chart</th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Namespace</th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Current Ver</th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Latest Ver</th>
                    <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-slate-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-slate-200">
                  {releases?.map((item: ComparisonResult) => (
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
          )}
        </div>
      </main>
    </div>
  )
}

export default App
