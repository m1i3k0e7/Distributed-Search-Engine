
import React, { useState, useEffect } from 'react';
import { useSearchParams, useNavigate, Link as RouterLink } from 'react-router-dom';
import SearchBar from '../components/SearchBar';
import Results from '../components/Results';
import FilterSidebar from '../components/FilterSidebar';
import { mockProducts } from '../mock-data';

import { Typography, CircularProgress, Box, Button } from '@mui/material';
import HomeIcon from '@mui/icons-material/Home';

function ResultsPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const query = searchParams.get('q') || '';
  const initialCategories = searchParams.getAll('cat') || [];

  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedCategories, setSelectedCategories] = useState(initialCategories);

  const handleCategoryChange = (newCategories) => {
    const params = new URLSearchParams(searchParams);
    params.delete('cat');
    newCategories.forEach(cat => params.append('cat', cat));
    navigate(`/search?${params.toString()}`, { replace: true });
  };

  const handleSearch = (newQuery) => {
    if (!newQuery.trim()) return;
    const params = new URLSearchParams(searchParams);
    params.set('q', newQuery.trim());
    navigate(`/search?${params.toString()}`);
  };

  useEffect(() => {
    const currentQuery = searchParams.get('q') || '';
    const currentCategories = searchParams.getAll('cat') || [];

    const fetchResults = async () => {
      setLoading(true);
      setError(null);
      setSelectedCategories(currentCategories);

      if (currentQuery === 'mock') {
        setResults(mockProducts);
        setLoading(false);
        return;
      }

      if (!currentQuery) {
        setResults([]);
        setLoading(false);
        return;
      }

      try {
        const response = await fetch('http://127.0.0.1:5678/search', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            Query: currentQuery,
            // Keywords: currentQuery.split(' ').filter(kw => kw),
            Classes: currentCategories,
            // PirceFrom: 0,
            // PriceTo: 10000,
          }),
        });

        if (!response.ok) {
          const errorBody = await response.text();
          throw new Error(`HTTP error! status: ${response.status}, body: ${errorBody}`);
        }

        const data = await response.json();
        setResults(data || []);
      } catch (e) {
        setError(`Failed to fetch results: ${e.message}`);
        console.error(e);
        setResults([]);
      } finally {
        setLoading(false);
      }
    };

    fetchResults();
  }, [searchParams.toString()]);

  return (
    <Box sx={{ my: 2 }}>
      <Box> {/* New wrapper Box */}
        <Typography variant="h4" component="h1" align="center" gutterBottom>
          Nexus Search
        </Typography>
        <Box sx={{ display: 'flex', flexDirection: { xs: 'column', md: 'row' }, gap: 4, my: 2 }}>
          {/* Sidebar */}
          <Box sx={{ width: { xs: '100%', md: '25%' }, minWidth: { md: '280px' }, flexShrink: { md: 0 } }}>
            <FilterSidebar
              selectedCategories={selectedCategories}
              onCategoryChange={handleCategoryChange}
            />
          </Box>

          {/* Main Content */}
          <Box sx={{ flexGrow: 1 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
              <Button component={RouterLink} to="/" variant="outlined" startIcon={<HomeIcon />}>
                Home
              </Button>
              <SearchBar initialQuery={query} onSearch={handleSearch} />
            </Box>
            <Box sx={{ mt: 2 }}>
              {loading ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', my: 4 }}>
                  <CircularProgress size={60} />
                </Box>
              ) : error ? (
                <Typography color="error" align="center">{error}</Typography>
              ) : (
                <Results products={results} />
              )}
            </Box>
          </Box>
        </Box> {/* Closing New wrapper Box */}
      </Box>
    </Box>
  );
}

export default ResultsPage;
