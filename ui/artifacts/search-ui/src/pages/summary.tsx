import { useEffect, useState } from "react";
import { Link } from "wouter";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from "recharts";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";

interface SummaryData {
  SearchType: string;
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

  const chartData = data.map(d => ({
    type: d.SearchType,
    precision: d.Total > 0 ? (d.Relevant / d.Total) * 100 : 0
  }));

  return (
    <div className="min-h-screen container mx-auto p-10">
      <header className="mb-10 flex justify-between items-center">
        <h1 className="text-3xl font-bold">Evaluation Summary</h1>
        <Link href="/" className="text-primary hover:underline">Back to Home</Link>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-10">
        <div className="h-80 border rounded p-4">
          <h2 className="text-xl mb-4">Precision (%)</h2>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="type" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="precision" fill="#8884d8" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <Table className="border rounded">
          <TableHeader>
            <TableRow>
              <TableHead>Type</TableHead>
              <TableHead>Total Queries</TableHead>
              <TableHead>Relevant</TableHead>
              <TableHead>Precision</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.map(d => (
              <TableRow key={d.SearchType}>
                <TableCell className="font-medium">{d.SearchType}</TableCell>
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
