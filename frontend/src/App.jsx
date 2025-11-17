import React, { useState, useEffect, useRef } from 'react';
import { TrendingUp, TrendingDown, DollarSign, Activity, BarChart3 } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

// const API_URL = 'http://localhost:8080';
// const WS_URL = 'ws://localhost:8080/ws';
// for AWS 
const API_URL = 'http://51.21.219.168:8080';
const WS_URL = 'ws://51.21.219.168:8080/ws';

export default function TradingDashboard() {
  const [prices, setPrices] = useState({});
  const [orders, setOrders] = useState([]);
  const [orderForm, setOrderForm] = useState({
    symbol: 'AAPL',
    side: 'buy',
    quantity: 1,
    price: 0
  });
  const [wsStatus, setWsStatus] = useState('connecting');
  const [notification, setNotification] = useState(null);
  const [priceHistory, setPriceHistory] = useState({});
  const [selectedChart, setSelectedChart] = useState('AAPL');
  const wsRef = useRef(null);
  const previousPricesRef = useRef({});

  // WebSocket connection
  useEffect(() => {
    connectWebSocket();
    fetchOrders();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const connectWebSocket = () => {
    const ws = new WebSocket(WS_URL);
    
    ws.onopen = () => {
      setWsStatus('connected');
      showNotification('Connected to market data', 'success');
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      previousPricesRef.current = { ...prices };
      setPrices(data);
      
      // Update price history for charts
      const timestamp = new Date().toLocaleTimeString();
      setPriceHistory(prev => {
        const updated = { ...prev };
        Object.keys(data).forEach(symbol => {
          if (!updated[symbol]) {
            updated[symbol] = [];
          }
          updated[symbol] = [
            ...updated[symbol].slice(-29), // Keep last 30 data points
            {
              time: timestamp,
              price: data[symbol].price,
              change: data[symbol].change
            }
          ];
        });
        return updated;
      });
    };

    ws.onerror = () => {
      setWsStatus('error');
      showNotification('WebSocket error', 'error');
    };

    ws.onclose = () => {
      setWsStatus('disconnected');
      setTimeout(connectWebSocket, 3000);
    };

    wsRef.current = ws;
  };

  const fetchOrders = async () => {
    try {
      const response = await fetch(`${API_URL}/orders`);
      const data = await response.json();
      setOrders(data || []);
    } catch (error) {
      console.error('Error fetching orders:', error);
    }
  };

  const handleSubmitOrder = async () => {
    try {
      const response = await fetch(`${API_URL}/orders`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...orderForm,
          quantity: parseInt(orderForm.quantity),
          price: parseFloat(orderForm.price)
        })
      });

      if (response.ok) {
        const newOrder = await response.json();
        setOrders([newOrder, ...orders]);
        showNotification(`Order placed: ${orderForm.side.toUpperCase()} ${orderForm.quantity} ${orderForm.symbol}`, 'success');
        
        setOrderForm({
          ...orderForm,
          quantity: 1,
          price: prices[orderForm.symbol]?.price || 0
        });
      } else {
        showNotification('Failed to place order', 'error');
      }
    } catch (error) {
      showNotification('Error placing order', 'error');
    }
  };

  const showNotification = (message, type) => {
    setNotification({ message, type });
    setTimeout(() => setNotification(null), 3000);
  };

  const getPriceChangeColor = (symbol) => {
    const current = prices[symbol];
    if (!current || current.change === 0) return 'text-gray-400';
    return current.change > 0 ? 'text-green-500' : 'text-red-500';
  };

  const formatPrice = (price) => {
    return price?.toFixed(2) || '0.00';
  };

  const formatChange = (change) => {
    if (!change) return '0.00%';
    const sign = change > 0 ? '+' : '';
    return `${sign}${change.toFixed(2)}%`;
  };

  useEffect(() => {
    if (prices[orderForm.symbol]) {
      setOrderForm(prev => ({
        ...prev,
        price: prices[orderForm.symbol].price
      }));
    }
  }, [orderForm.symbol, prices]);

  const CustomTooltip = ({ active, payload }) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-slate-800 border border-slate-600 rounded-lg p-3 shadow-lg">
          <p className="text-gray-400 text-sm">{payload[0].payload.time}</p>
          <p className="text-white font-semibold">${payload[0].value.toFixed(2)}</p>
          <p className={`text-sm ${payload[0].payload.change > 0 ? 'text-green-400' : 'text-red-400'}`}>
            {formatChange(payload[0].payload.change)}
          </p>
        </div>
      );
    }
    return null;
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Activity className="w-8 h-8 text-blue-400" />
              <h1 className="text-3xl font-bold text-white">Trading Dashboard</h1>
            </div>
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${
                wsStatus === 'connected' ? 'bg-green-500' : 
                wsStatus === 'connecting' ? 'bg-yellow-500' : 'bg-red-500'
              } animate-pulse`}></div>
              <span className="text-sm text-gray-400 capitalize">{wsStatus}</span>
            </div>
          </div>
        </div>

        {/* Notification */}
        {notification && (
          <div className={`mb-4 p-4 rounded-lg ${
            notification.type === 'success' ? 'bg-green-500/20 border border-green-500' : 
            'bg-red-500/20 border border-red-500'
          }`}>
            <p className={notification.type === 'success' ? 'text-green-400' : 'text-red-400'}>
              {notification.message}
            </p>
          </div>
        )}

        {/* Live Chart */}
        <div className="mb-8 bg-slate-800/50 backdrop-blur rounded-lg border border-slate-700 overflow-hidden">
          <div className="p-4 border-b border-slate-700 flex items-center justify-between">
            <h2 className="text-xl font-semibold text-white flex items-center gap-2">
              <BarChart3 className="w-5 h-5 text-blue-400" />
              Live Price Chart
            </h2>
            <div className="flex gap-2">
              {Object.keys(prices).map(symbol => (
                <button
                  key={symbol}
                  onClick={() => setSelectedChart(symbol)}
                  className={`px-4 py-2 rounded-lg text-sm font-semibold transition-colors ${
                    selectedChart === symbol
                      ? 'bg-blue-600 text-white'
                      : 'bg-slate-700 text-gray-400 hover:bg-slate-600'
                  }`}
                >
                  {symbol}
                </button>
              ))}
            </div>
          </div>
          <div className="p-6">
            {priceHistory[selectedChart] && priceHistory[selectedChart].length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={priceHistory[selectedChart]}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                  <XAxis 
                    dataKey="time" 
                    stroke="#9CA3AF"
                    tick={{ fill: '#9CA3AF', fontSize: 12 }}
                  />
                  <YAxis 
                    stroke="#9CA3AF"
                    tick={{ fill: '#9CA3AF', fontSize: 12 }}
                    domain={['auto', 'auto']}
                    tickFormatter={(value) => `$${value.toFixed(2)}`}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Line 
                    type="monotone" 
                    dataKey="price" 
                    stroke="#3B82F6" 
                    strokeWidth={2}
                    dot={false}
                    isAnimationActive={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[300px] flex items-center justify-center text-gray-500">
                Waiting for price data...
              </div>
            )}
          </div>
        </div>

        {/* Live Prices */}
        <div className="mb-8 bg-slate-800/50 backdrop-blur rounded-lg border border-slate-700 overflow-hidden">
          <div className="p-4 border-b border-slate-700">
            <h2 className="text-xl font-semibold text-white flex items-center gap-2">
              <DollarSign className="w-5 h-5 text-blue-400" />
              Live Market Prices
            </h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="bg-slate-900/50">
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Symbol</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-400 uppercase tracking-wider">Price</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-400 uppercase tracking-wider">Change</th>
                  <th className="px-6 py-3 text-center text-xs font-medium text-gray-400 uppercase tracking-wider">Trend</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {Object.values(prices).map((stock) => (
                  <tr key={stock.symbol} className="hover:bg-slate-700/30 transition-colors">
                    <td className="px-6 py-4">
                      <span className="text-white font-semibold">{stock.symbol}</span>
                    </td>
                    <td className={`px-6 py-4 text-right font-mono text-lg ${getPriceChangeColor(stock.symbol)}`}>
                      ${formatPrice(stock.price)}
                    </td>
                    <td className={`px-6 py-4 text-right font-mono ${getPriceChangeColor(stock.symbol)}`}>
                      {formatChange(stock.change)}
                    </td>
                    <td className="px-6 py-4 text-center">
                      {stock.change > 0 ? (
                        <TrendingUp className="w-5 h-5 text-green-500 inline" />
                      ) : stock.change < 0 ? (
                        <TrendingDown className="w-5 h-5 text-red-500 inline" />
                      ) : (
                        <div className="w-5 h-5 inline-block" />
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Order Form */}
        <div className="mb-8 bg-slate-800/50 backdrop-blur rounded-lg border border-slate-700 p-6">
          <h2 className="text-xl font-semibold text-white mb-4">Place Order</h2>
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">Symbol</label>
              <select
                value={orderForm.symbol}
                onChange={(e) => setOrderForm({ ...orderForm, symbol: e.target.value })}
                className="w-full bg-slate-900 border border-slate-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {Object.keys(prices).map(symbol => (
                  <option key={symbol} value={symbol}>{symbol}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">Side</label>
              <select
                value={orderForm.side}
                onChange={(e) => setOrderForm({ ...orderForm, side: e.target.value })}
                className="w-full bg-slate-900 border border-slate-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="buy">Buy</option>
                <option value="sell">Sell</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">Quantity</label>
              <input
                type="number"
                min="1"
                value={orderForm.quantity}
                onChange={(e) => setOrderForm({ ...orderForm, quantity: e.target.value })}
                className="w-full bg-slate-900 border border-slate-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">Price ($)</label>
              <input
                type="number"
                step="0.01"
                min="0.01"
                value={orderForm.price}
                onChange={(e) => setOrderForm({ ...orderForm, price: e.target.value })}
                className="w-full bg-slate-900 border border-slate-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div className="flex items-end">
              <button
                onClick={handleSubmitOrder}
                className={`w-full py-2 px-4 rounded-lg font-semibold transition-colors ${
                  orderForm.side === 'buy'
                    ? 'bg-green-600 hover:bg-green-700 text-white'
                    : 'bg-red-600 hover:bg-red-700 text-white'
                }`}
              >
                {orderForm.side === 'buy' ? 'Place Buy Order' : 'Place Sell Order'}
              </button>
            </div>
          </div>
        </div>

        {/* Orders Table */}
        <div className="bg-slate-800/50 backdrop-blur rounded-lg border border-slate-700 overflow-hidden">
          <div className="p-4 border-b border-slate-700">
            <h2 className="text-xl font-semibold text-white">Order History</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="bg-slate-900/50">
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">ID</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Time</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Symbol</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Side</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-400 uppercase tracking-wider">Quantity</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-400 uppercase tracking-wider">Price</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-400 uppercase tracking-wider">Total</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {orders.length === 0 ? (
                  <tr>
                    <td colSpan="7" className="px-6 py-8 text-center text-gray-500">
                      No orders placed yet
                    </td>
                  </tr>
                ) : (
                  orders.map((order) => (
                    <tr key={order.id} className="hover:bg-slate-700/30 transition-colors">
                      <td className="px-6 py-4 text-white">#{order.id}</td>
                      <td className="px-6 py-4 text-gray-400 text-sm">{order.time}</td>
                      <td className="px-6 py-4 text-white font-semibold">{order.symbol}</td>
                      <td className="px-6 py-4">
                        <span className={`px-3 py-1 rounded-full text-xs font-semibold ${
                          order.side === 'buy' 
                            ? 'bg-green-500/20 text-green-400 border border-green-500/50' 
                            : 'bg-red-500/20 text-red-400 border border-red-500/50'
                        }`}>
                          {order.side.toUpperCase()}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right text-white font-mono">{order.quantity}</td>
                      <td className="px-6 py-4 text-right text-white font-mono">${formatPrice(order.price)}</td>
                      <td className="px-6 py-4 text-right text-white font-mono font-semibold">
                        ${formatPrice(order.quantity * order.price)}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}