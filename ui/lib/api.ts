const API_BASE = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080").replace(/\/$/, "");

export interface IsotopeInfo {
  name: string;
  fission_xs: number;
  absorption_xs: number;
  scatter_xs: number;
  nu: number;
  density: number;
  description: string;
}

export interface SimRequest {
  isotope: string;
  thickness: number;
  geometry: string;
  particles: number;
  generations: number;
  seed: number;
}

export interface SimResponse {
  material: string;
  geometry: string;
  thickness: number;
  num_particles: number;
  num_generations: number;
  seed: number;
  total_fissions: number;
  total_absorptions: number;
  total_escapes: number;
  k_effective: number;
  status: string;
  neutrons_per_gen: number[];
  mean_free_path: number;
  sigma_total: number;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const error = await res.text();
    throw new Error(error || `HTTP ${res.status}`);
  }

  return res.json();
}

export const api = {
  health: () => request<{ status: string }>("/api/health"),

  getIsotopes: () =>
    request<{ isotopes: IsotopeInfo[] }>("/api/isotopes"),

  simulate: (data: SimRequest) =>
    request<SimResponse>("/api/simulate", {
      method: "POST",
      body: JSON.stringify(data),
    }),
};
