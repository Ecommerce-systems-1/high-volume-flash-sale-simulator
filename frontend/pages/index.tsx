import { useState, useEffect, useCallback } from 'react'

interface StatsResponse {
  sale_id: number
  product_name: string
  initial_stock: number
  stock_remaining: number
  total_requests: number
  successful: number
  rejected_sold_out: number
  rejected_expired: number
  rps: number
  sale_active: boolean
  seconds_remaining: number
}

export default function Home() {
  const [stats, setStats] = useState<StatsResponse | null>(null)
  const [userId, setUserId] = useState('')
  const [reserving, setReserving] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    setUserId(`user_${Math.floor(Math.random() * 100000)}`)
  }, [])

  const fetchStats = useCallback(async () => {
    try {
      const res = await fetch('/api/stats?sale_id=1')
      if (res.ok) {
        const data = await res.json()
        setStats(data)
      }
    } catch {
      // ignore polling errors
    }
  }, [])

  useEffect(() => {
    fetchStats()
    const interval = setInterval(fetchStats, 2000)
    return () => clearInterval(interval)
  }, [fetchStats])

  const handleReserve = async () => {
    setReserving(true)
    setMessage('')
    try {
      const res = await fetch('/api/reserve', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: userId, sale_id: 1 }),
      })
      const data = await res.json()
      if (res.ok) {
        setMessage(`✅ Reserved! Order: ${data.order_id.substring(0, 8)}...`)
        fetchStats()
      } else if (res.status === 409) {
        setMessage('❌ Sold out! No stock remaining.')
      } else if (res.status === 410) {
        setMessage('⏰ Sale has ended!')
      } else {
        setMessage(`❌ Error: ${data.error || 'unknown'}`)
      }
    } catch {
      setMessage('❌ Network error')
    } finally {
      setReserving(false)
    }
  }

  const formatTime = (seconds: number) => {
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return `${m}:${s.toString().padStart(2, '0')}`
  }

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100 flex flex-col items-center p-4">
      <h1 className="text-3xl font-bold mt-8 mb-2">⚡ Flash Sale Simulator</h1>
      <p className="text-gray-400 mb-8">High-volume atomic inventory reservation demo</p>

      {stats && (
        <div className="w-full max-w-md space-y-4">
          {/* Countdown */}
          <div className="bg-gray-900 rounded-lg p-6 text-center border border-gray-800">
            <p className="text-sm text-gray-400 mb-2">{stats.product_name}</p>
            {stats.sale_active ? (
              <>
                <p className="text-5xl font-mono font-bold text-emerald-400">
                  {formatTime(stats.seconds_remaining)}
                </p>
                <p className="text-sm text-gray-400 mt-2">remaining</p>
              </>
            ) : (
              <p className="text-2xl font-bold text-red-400">SALE ENDED</p>
            )}
          </div>

          {/* Stock */}
          <div className="bg-gray-900 rounded-lg p-6 border border-gray-800">
            <div className="flex justify-between mb-2">
              <span className="text-gray-400">Stock Available</span>
              <span className={`text-2xl font-bold ${stats.stock_remaining > 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                {stats.stock_remaining}
              </span>
            </div>
            <div className="w-full bg-gray-800 rounded-full h-3">
              <div
                className={`h-3 rounded-full transition-all duration-500 ${stats.stock_remaining > 0 ? 'bg-emerald-500' : 'bg-red-500'}`}
                style={{ width: `${(stats.stock_remaining / stats.initial_stock) * 100}%` }}
              />
            </div>
            <p className="text-xs text-gray-500 mt-1">
              of {stats.initial_stock} initial
            </p>
          </div>

          {/* Stats */}
          <div className="bg-gray-900 rounded-lg p-6 border border-gray-800">
            <h2 className="text-sm font-semibold text-gray-400 mb-3">Live Stats</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-gray-500">Total Requests</p>
                <p className="text-xl font-bold">{stats.total_requests}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500">Successful</p>
                <p className="text-xl font-bold text-emerald-400">{stats.successful}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500">Sold Out</p>
                <p className="text-xl font-bold text-red-400">{stats.rejected_sold_out}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500">Throughput</p>
                <p className="text-xl font-bold">{stats.rps.toFixed(1)} RPS</p>
              </div>
            </div>
          </div>

          {/* Reserve Button */}
          <button
            onClick={handleReserve}
            disabled={reserving || !stats.sale_active || stats.stock_remaining <= 0}
            className="w-full py-4 rounded-lg font-bold text-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700"
          >
            {reserving ? 'Reserving...' : stats.stock_remaining <= 0 ? 'Sold Out' : '⚡ Reserve Now'}
          </button>

          {message && (
            <div className="bg-gray-900 rounded-lg p-4 text-center border border-gray-800 text-sm">
              {message}
            </div>
          )}

          {/* User ID */}
          <div className="text-center text-xs text-gray-600">
            User: {userId}
          </div>
        </div>
      )}

      {!stats && (
        <div className="text-gray-500 mt-20">
          <p className="animate-pulse">Loading flash sale data...</p>
        </div>
      )}
    </div>
  )
}