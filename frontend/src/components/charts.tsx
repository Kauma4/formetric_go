"use client"

import { useEffect, useState } from "react"
import {
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  LineChart as RechartsLineChart,
  Line,
  CartesianGrid,
  ResponsiveContainer
} from "recharts"

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8']

// Объединённый график
export function CombinedChart({ data }: { data: { name: string; value: number }[] }) {
  return (
    <div className="space-y-6">
      {/* Гистограмма */}
      <div>
        <h3 className="text-xl font-semibold mb-4">Гистограмма</h3>
        <ResponsiveContainer width="100%" height={300}>
          <RechartsBarChart data={data}>
            <XAxis dataKey="name" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="value" fill="#8884d8" />
          </RechartsBarChart>
        </ResponsiveContainer>
      </div>

      {/* Круговая диаграмма */}
      <div>
        <h3 className="text-xl font-semibold mb-4">Круговая диаграмма</h3>
        <ResponsiveContainer width="100%" height={300}>
          <RechartsPieChart>
            <Pie
              data={data}
              cx="50%"
              cy="50%"
              labelLine={false}
              outerRadius={150}
              fill="#8884d8"
              dataKey="value"
              label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
            >
              {data.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip />
            <Legend />
          </RechartsPieChart>
        </ResponsiveContainer>
      </div>

      {/* Линейный график */}
      <div>
        <h3 className="text-xl font-semibold mb-4">Линейный график</h3>
        <ResponsiveContainer width="100%" height={300}>
          <RechartsLineChart data={data} margin={{ top: 5, right: 5, left: 5, bottom: 5 }}>
            <XAxis dataKey="name" />
            <YAxis />
            <Tooltip />
            <Legend />
            <CartesianGrid strokeDasharray="3 3" />
            <Line type="monotone" dataKey="value" stroke="#8884d8" animationDuration={500} />
          </RechartsLineChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}

/*export function AppBarChart({ data }: { data: { name: string; value: number }[] }) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <RechartsBarChart data={data}>
        <XAxis dataKey="name" />
        <YAxis />
        <Tooltip />
        <Legend />
        <RechartsBar dataKey="value" fill="#8884d8" />
      </RechartsBarChart>
    </ResponsiveContainer>
  )
}

export function AppPieChart({ data }: { data: { name: string; value: number }[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <RechartsPieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          labelLine={false}
          outerRadius={150}
          fill="#8884d8"
          dataKey="value"
          label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
        >
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip />
        <Legend />
      </RechartsPieChart>
    </ResponsiveContainer>
  )
}*/
