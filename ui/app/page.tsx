"use client";

import { useEffect, useState, useRef } from "react";
import { api, SimResponse, IsotopeInfo } from "../lib/api";

export default function Home() {
  const [isotopes, setIsotopes] = useState<IsotopeInfo[]>([]);
  const [isotope, setIsotope] = useState("U-235");
  const [thickness, setThickness] = useState(20);
  const [geometry, setGeometry] = useState("slab");
  const [particles, setParticles] = useState(1000);
  const [generations, setGenerations] = useState(50);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<SimResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    api.getIsotopes().then((data) => setIsotopes(data.isotopes)).catch(() => {});
  }, []);

  useEffect(() => {
    if (result && canvasRef.current) {
      drawChart(canvasRef.current, result.neutrons_per_gen);
    }
  }, [result]);

  const handleRun = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.simulate({
        isotope,
        thickness,
        geometry,
        particles,
        generations,
        seed: 0,
      });
      setResult(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Simulation failed");
    } finally {
      setLoading(false);
    }
  };

  const selectedIsotope = isotopes.find((i) => i.name === isotope);

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-indigo-950 text-white">
      {/* Header */}
      <header className="border-b border-slate-800/50 backdrop-blur-xl bg-slate-950/50 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-orange-500 to-red-600 flex items-center justify-center shadow-lg shadow-orange-500/25">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold">MAKAROV</h1>
              <p className="text-xs text-slate-400">Monte Carlo Chain Reaction Simulator</p>
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-slate-500">
            <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></div>
            ENDF/B-VIII.0 Data
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left: Controls */}
          <div className="space-y-6">
            {/* Material Selection */}
            <div className="bg-slate-900/50 border border-slate-800/50 rounded-2xl p-6">
              <h2 className="text-sm font-bold text-slate-400 uppercase tracking-wider mb-4">Material</h2>
              <select
                value={isotope}
                onChange={(e) => setIsotope(e.target.value)}
                className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl px-4 py-3 text-white font-medium focus:outline-none focus:ring-2 focus:ring-orange-500/50"
              >
                {isotopes.map((iso) => (
                  <option key={iso.name} value={iso.name}>
                    {iso.name} — {iso.description}
                  </option>
                ))}
              </select>

              {selectedIsotope && (
                <div className="mt-4 grid grid-cols-2 gap-3">
                  <div className="bg-slate-800/30 rounded-xl p-3">
                    <p className="text-xs text-slate-500">Fission XS</p>
                    <p className="text-sm font-bold text-orange-400">{selectedIsotope.fission_xs} b</p>
                  </div>
                  <div className="bg-slate-800/30 rounded-xl p-3">
                    <p className="text-xs text-slate-500">Capture XS</p>
                    <p className="text-sm font-bold text-blue-400">{selectedIsotope.absorption_xs} b</p>
                  </div>
                  <div className="bg-slate-800/30 rounded-xl p-3">
                    <p className="text-xs text-slate-500">Scatter XS</p>
                    <p className="text-sm font-bold text-green-400">{selectedIsotope.scatter_xs} b</p>
                  </div>
                  <div className="bg-slate-800/30 rounded-xl p-3">
                    <p className="text-xs text-slate-500">ν (neutrons/fission)</p>
                    <p className="text-sm font-bold text-purple-400">{selectedIsotope.nu}</p>
                  </div>
                </div>
              )}
            </div>

            {/* Geometry */}
            <div className="bg-slate-900/50 border border-slate-800/50 rounded-2xl p-6">
              <h2 className="text-sm font-bold text-slate-400 uppercase tracking-wider mb-4">Geometry</h2>
              <div className="grid grid-cols-2 gap-3">
                <button
                  onClick={() => setGeometry("slab")}
                  className={`p-4 rounded-xl border-2 transition-all ${
                    geometry === "slab"
                      ? "border-orange-500 bg-orange-500/10 text-orange-400"
                      : "border-slate-700/50 bg-slate-800/30 text-slate-400 hover:border-slate-600"
                  }`}
                >
                  <div className="text-2xl mb-1">▬</div>
                  <div className="text-xs font-bold">1D Slab</div>
                </button>
                <button
                  onClick={() => setGeometry("sphere")}
                  className={`p-4 rounded-xl border-2 transition-all ${
                    geometry === "sphere"
                      ? "border-orange-500 bg-orange-500/10 text-orange-400"
                      : "border-slate-700/50 bg-slate-800/30 text-slate-400 hover:border-slate-600"
                  }`}
                >
                  <div className="text-2xl mb-1">●</div>
                  <div className="text-xs font-bold">3D Sphere</div>
                </button>
              </div>

              <div className="mt-4">
                <label className="text-xs text-slate-500 mb-1 block">
                  {geometry === "slab" ? "Thickness" : "Radius"} (cm)
                </label>
                <input
                  type="number"
                  value={thickness}
                  onChange={(e) => setThickness(Number(e.target.value))}
                  min={1}
                  max={100}
                  className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl px-4 py-2 text-white font-mono focus:outline-none focus:ring-2 focus:ring-orange-500/50"
                />
              </div>
            </div>

            {/* Simulation Parameters */}
            <div className="bg-slate-900/50 border border-slate-800/50 rounded-2xl p-6">
              <h2 className="text-sm font-bold text-slate-400 uppercase tracking-wider mb-4">Parameters</h2>
              <div className="space-y-4">
                <div>
                  <label className="text-xs text-slate-500 mb-1 block">Neutrons per generation</label>
                  <input
                    type="number"
                    value={particles}
                    onChange={(e) => setParticles(Number(e.target.value))}
                    min={100}
                    max={100000}
                    step={100}
                    className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl px-4 py-2 text-white font-mono focus:outline-none focus:ring-2 focus:ring-orange-500/50"
                  />
                </div>
                <div>
                  <label className="text-xs text-slate-500 mb-1 block">Generations</label>
                  <input
                    type="number"
                    value={generations}
                    onChange={(e) => setGenerations(Number(e.target.value))}
                    min={10}
                    max={500}
                    step={10}
                    className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl px-4 py-2 text-white font-mono focus:outline-none focus:ring-2 focus:ring-orange-500/50"
                  />
                </div>
              </div>
            </div>

            {/* Run Button */}
            <button
              onClick={handleRun}
              disabled={loading}
              className="w-full py-4 rounded-2xl bg-gradient-to-r from-orange-500 to-red-600 text-white font-bold text-lg shadow-lg shadow-orange-500/25 hover:shadow-orange-500/40 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Simulating...
                </span>
              ) : (
                "Run Simulation"
              )}
            </button>
          </div>

          {/* Right: Results */}
          <div className="lg:col-span-2 space-y-6">
            {error && (
              <div className="bg-red-500/10 border border-red-500/30 rounded-2xl p-4 text-red-400 text-sm">
                {error}
              </div>
            )}

            {result && (
              <>
                {/* Status Banner */}
                <div
                  className={`rounded-2xl p-6 border-2 ${
                    result.status === "supercritical"
                      ? "bg-red-500/10 border-red-500/30"
                      : result.status === "subcritical"
                      ? "bg-blue-500/10 border-blue-500/30"
                      : "bg-yellow-500/10 border-yellow-500/30"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <h2 className="text-3xl font-bold">
                        k<sub>eff</sub> = {result.k_effective.toFixed(4)}
                      </h2>
                      <p
                        className={`text-lg font-medium mt-1 ${
                          result.status === "supercritical"
                            ? "text-red-400"
                            : result.status === "subcritical"
                            ? "text-blue-400"
                            : "text-yellow-400"
                        }`}
                      >
                        {result.status === "supercritical"
                          ? "SUPERCRITICAL — Growing reaction"
                          : result.status === "subcritical"
                          ? "SUBCRITICAL — Dying reaction"
                          : "CRITICAL — Sustained reaction"}
                      </p>
                    </div>
                    <div className="text-right text-sm text-slate-400">
                      <p>{result.material} • {result.geometry}</p>
                      <p>{result.thickness} cm • {result.num_particles} neutrons</p>
                    </div>
                  </div>
                </div>

                {/* Stats Grid */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <StatCard label="Fissions" value={result.total_fissions.toLocaleString()} color="text-orange-400" />
                  <StatCard label="Absorptions" value={result.total_absorptions.toLocaleString()} color="text-blue-400" />
                  <StatCard label="Escapes" value={result.total_escapes.toLocaleString()} color="text-green-400" />
                  <StatCard label="MFP" value={`${result.mean_free_path.toFixed(4)} cm`} color="text-purple-400" />
                </div>

                {/* Chart */}
                <div className="bg-slate-900/50 border border-slate-800/50 rounded-2xl p-6">
                  <h3 className="text-sm font-bold text-slate-400 uppercase tracking-wider mb-4">
                    Neutron Population per Generation
                  </h3>
                  <canvas ref={canvasRef} width={800} height={300} className="w-full" />
                </div>
              </>
            )}

            {!result && !loading && (
              <div className="bg-slate-900/50 border border-slate-800/50 rounded-2xl p-12 text-center">
                <div className="text-6xl mb-4">⚛️</div>
                <h2 className="text-xl font-bold text-slate-300 mb-2">Configure & Run</h2>
                <p className="text-slate-500">
                  Select a material, set geometry and parameters, then run the simulation.
                </p>
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: string; color: string }) {
  return (
    <div className="bg-slate-900/50 border border-slate-800/50 rounded-xl p-4">
      <p className="text-xs text-slate-500 mb-1">{label}</p>
      <p className={`text-lg font-bold font-mono ${color}`}>{value}</p>
    </div>
  );
}

function drawChart(canvas: HTMLCanvasElement, data: number[]) {
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  const w = canvas.width;
  const h = canvas.height;
  const padding = { top: 20, right: 20, bottom: 40, left: 60 };

  ctx.clearRect(0, 0, w, h);

  const maxVal = Math.max(...data) * 1.1;
  const minVal = 0;
  const xScale = (w - padding.left - padding.right) / (data.length - 1 || 1);
  const yScale = (h - padding.top - padding.bottom) / (maxVal - minVal);

  // Grid
  ctx.strokeStyle = "#334155";
  ctx.lineWidth = 0.5;
  for (let i = 0; i <= 5; i++) {
    const y = padding.top + (i / 5) * (h - padding.top - padding.bottom);
    ctx.beginPath();
    ctx.moveTo(padding.left, y);
    ctx.lineTo(w - padding.right, y);
    ctx.stroke();

    ctx.fillStyle = "#64748b";
    ctx.font = "11px monospace";
    ctx.textAlign = "right";
    ctx.fillText(Math.round(maxVal - (i / 5) * maxVal).toString(), padding.left - 8, y + 4);
  }

  // X axis labels
  ctx.fillStyle = "#64748b";
  ctx.textAlign = "center";
  for (let i = 0; i < data.length; i += Math.max(1, Math.floor(data.length / 10))) {
    const x = padding.left + i * xScale;
    ctx.fillText((i + 1).toString(), x, h - 10);
  }

  // Line
  ctx.beginPath();
  ctx.strokeStyle = "#f97316";
  ctx.lineWidth = 2;
  ctx.lineJoin = "round";

  for (let i = 0; i < data.length; i++) {
    const x = padding.left + i * xScale;
    const y = padding.top + (maxVal - data[i]) * yScale;
    if (i === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  }
  ctx.stroke();

  // Fill under line
  ctx.lineTo(padding.left + (data.length - 1) * xScale, h - padding.bottom);
  ctx.lineTo(padding.left, h - padding.bottom);
  ctx.closePath();
  ctx.fillStyle = "rgba(249, 115, 22, 0.1)";
  ctx.fill();

  // Data points
  for (let i = 0; i < data.length; i++) {
    const x = padding.left + i * xScale;
    const y = padding.top + (maxVal - data[i]) * yScale;

    ctx.beginPath();
    ctx.arc(x, y, 3, 0, Math.PI * 2);
    ctx.fillStyle = "#f97316";
    ctx.fill();
  }
}
