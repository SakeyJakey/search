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

  const volumeChartData = categories.map(cat => {
    const catData = data.filter(d => d.Category === cat);
    const entry: any = { category: cat };
    catData.forEach(d => {
      entry[`${d.SearchType}_Total`] = d.Total;
      entry[`${d.SearchType}_Relevant`] = d.Relevant;
    });
    return entry;
  });

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
          <h2 className="text-xl mb-4">Query Volume by Category</h2>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={volumeChartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="category" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="vector_Total" fill="#8884d8" name="Vector Total" />
              <Bar dataKey="vector_Relevant" fill="#8884d8" name="Vector Relevant" />
              <Bar dataKey="tfidf_Total" fill="#82ca9d" name="TF-IDF Total" />
              <Bar dataKey="tfidf_Relevant" fill="#82ca9d" name="TF-IDF Relevant" />
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
