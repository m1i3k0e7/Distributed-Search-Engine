import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import SearchBar from '../components/SearchBar';
import { CATEGORIES } from '../constants';
import {
  Box,
  Typography,
  Chip,
  Stack,
} from '@mui/material';

function HomePage() {
  const [selectedCategories, setSelectedCategories] = useState([]);
  const navigate = useNavigate();

  const handleCategoryClick = (category) => {
    setSelectedCategories((prev) => {
      if (prev.includes(category)) {
        return prev.filter((c) => c !== category);
      } else {
        return [...prev, category];
      }
    });
  };

  const handleSearch = (query) => {
    if (!query.trim()) return;

    const params = new URLSearchParams();
    params.append('q', query.trim());
    selectedCategories.forEach(cat => params.append('cat', cat));

    navigate(`/search?${params.toString()}`);
  };

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '90vh',
        textAlign: 'center',
        gap: 4,
      }}
    >
      <Typography variant="h1" color="text.primary">
        Nexus Search
      </Typography>
      <Typography variant="h6" color="text.secondary" sx={{ maxWidth: '600px' }}>
        Your one-stop destination to find and discover amazing products.
      </Typography>
      
      <Box sx={{ width: '100%', maxWidth: '700px' }}>
        <SearchBar onSearch={handleSearch} />
      </Box>

      <Box sx={{ maxWidth: '800px', width: '100%', p: 2 }}>
        <Stack direction="row" justifyContent="center" flexWrap="wrap" gap={1.5}>
          {CATEGORIES.map((category) => (
            <Chip
              key={category}
              label={category}
              variant={selectedCategories.includes(category) ? 'filled' : 'outlined'}
              onClick={() => handleCategoryClick(category)}
              color="primary"
            />
          ))}
        </Stack>
      </Box>
    </Box>
  );
}

export default HomePage;