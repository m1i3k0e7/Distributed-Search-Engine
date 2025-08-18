
import React, { useState, useEffect } from 'react';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import SearchIcon from '@mui/icons-material/Search';

function SearchBar({ initialQuery = '', onSearch }) {
  const [query, setQuery] = useState(initialQuery);
  
  useEffect(() => {
    setQuery(initialQuery);
  }, [initialQuery]);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (onSearch) {
      onSearch(query);
    }
  };

  return (
    <Box
      component="form"
      onSubmit={handleSubmit}
      sx={{ display: 'flex', alignItems: 'center', gap: 1, width: '100%' }}
    >
      <TextField
        fullWidth
        variant="outlined"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search for products, brands, and more..."
      />
      <Button
        type="submit"
        variant="contained"
        size="large"
        startIcon={<SearchIcon />}
        sx={{ height: '56px', flexShrink: 0 }}
      >
        Search
      </Button>
    </Box>
  );
}

export default SearchBar;
