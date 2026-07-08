import { useEffect, useState } from "react";
import { Link } from "wouter";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from "recharts";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";

interface SummaryData {
  SearchType: string;
  Category: string;
  Total: number;
  Relevant: number;
}

export default function SummaryPage() {
  const [data, setData] = useState<SummaryData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("http://localhost:9091/api/summary")
      .then(res => res.json())
      .then(data => {
        setData(data);
        setLoading(false);
      });
  }, []);

  if (loading) return <div className="p-10"><Skeleton className="h-60" /></div>;

  // Group by category
  const categories = Array.from(new Set(data.map(d => d.Category)));
  const chartData = categories.map(cat => {
    const catData = data.filter(d => d.Category === cat);
    const entry: any = { category: cat };
    catData.forEach(d => {
      entry[d.SearchType] = d.Total > 0 ? (d.Relevant / d.Total) * 100 : 0;
    });
    return entry;
  });

  // Calculate total precision
  const totalPrecision = data.reduce((acc, d) => {
    if (!acc[d.SearchType]) acc[d.SearchType] = { Total: 0, Relevant: 0 };
    acc[d.SearchType].Total += d.Total;
    acc[d.SearchType].Relevant += d.Relevant;
    return acc;
  }, {} as Record<string, { Total: number, Relevant: number }>);

  const totalPrecisionChartData = Object.entries(totalPrecision).map(([type, vals]) => ({
    type,
    precision: vals.Total > 0 ? (vals.Relevant / vals.Total) * 100 : 0
  }));

  // Helper to get fill color
  const getFillColor = (type: string) => type === 'tfidf' ? '#82ca9d' : '#8884d8';

  return (
    <div className="min-h-screen container mx-auto p-10">
      <header className="mb-10 flex justify-between items-center">
        <h1 className="text-3xl font-bold">Evaluation Summary</h1>
        <Link href="/" className="text-primary hover:underline">Back to Home</Link>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-10">
        <div className="h-80 border rounded p-4">
          <h2 className="text-xl mb-4">Precision (%) by Category</h2>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="category" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="vector" fill="#8884d8" name="Vector" />
              <Bar dataKey="tfidf" fill="#82ca9d" name="TF-IDF" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="h-80 border rounded p-4">
          <h2 className="text-xl mb-4">Total Precision (%)</h2>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={totalPrecisionChartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="type" />
              <YAxis />
              <Tooltip />
              {/* Removed Legend */}
              <Bar dataKey="precision" name="Precision (%)" shape={(props: any) => {
                  const { x, y, width, height, payload } = props;
                  return <rect x={x} y={y} width={width} height={height} fill={getFillColor(payload.type)} />;
              }} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <Table className="border rounded">
          <TableHeader>
            <TableRow>
              <TableHead>Category</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Total</TableHead>
              <TableHead>Relevant</TableHead>
              <TableHead>Precision</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.map(d => (
              <TableRow key={`${d.Category}-${d.SearchType}`}>
                <TableCell className="font-medium">{d.Category}</TableCell>
                <TableCell>{d.SearchType}</TableCell>
                <TableCell>{d.Total}</TableCell>
                <TableCell>{d.Relevant}</TableCell>
                <TableCell>{((d.Relevant / d.Total) * 100).toFixed(1)}%</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
