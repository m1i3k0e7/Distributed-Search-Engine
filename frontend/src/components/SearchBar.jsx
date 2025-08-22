
import React, { useState, useEffect, useRef } from 'react';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import SearchIcon from '@mui/icons-material/Search';
import Paper from '@mui/material/Paper';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';

function SearchBar({ initialQuery = '', onSearch, onSuggestionsHeightChange = () => {} }) {
  const [query, setQuery] = useState(initialQuery);
  const [suggestions, setSuggestions] = useState([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const debounceTimeout = useRef(null);
  const suggestionsRef = useRef(null);

  useEffect(() => {
    setQuery(initialQuery);
  }, [initialQuery]);

  useEffect(() => {
    if (showSuggestions && suggestionsRef.current) {
      onSuggestionsHeightChange(suggestionsRef.current.offsetHeight);
    } else {
      onSuggestionsHeightChange(0);
    }
  }, [suggestions, showSuggestions, onSuggestionsHeightChange]);

  const fetchSuggestions = async (searchQuery) => {
    if (!searchQuery.trim()) {
      setSuggestions([]);
      return;
    }
    try {
      const response = await fetch('http://127.0.0.1:5678/associate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ query: searchQuery }),
      });
      if (response.ok) {
        const data = await response.json();
        setSuggestions(data || []);
        setShowSuggestions(true);
      } else {
        setSuggestions([]);
      }
    } catch (error) {
      console.error('Error fetching suggestions:', error);
      setSuggestions([]);
    }
  };

  const handleInputChange = (e) => {
    const value = e.target.value;
    setQuery(value);

    if (debounceTimeout.current) {
      clearTimeout(debounceTimeout.current);
    }

    debounceTimeout.current = setTimeout(() => {
      fetchSuggestions(value);
    }, 300); // 300ms debounce time
  };

  const handleSuggestionClick = (suggestion) => {
    setQuery(suggestion);
    setSuggestions([]);
    setShowSuggestions(false);
    if (onSearch) {
      onSearch(suggestion);
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    setSuggestions([]);
    setShowSuggestions(false);
    if (onSearch) {
      onSearch(query);
    }
  };

  return (
    <Box
      component="form"
      onSubmit={handleSubmit}
      sx={{ position: 'relative', display: 'flex', alignItems: 'center', gap: 1, width: '100%' }}
    >
      <TextField
        fullWidth
        variant="outlined"
        value={query}
        onChange={handleInputChange}
        onBlur={() => setTimeout(() => setShowSuggestions(false), 150)} // Delay to allow click
        onFocus={() => query && suggestions.length > 0 && setShowSuggestions(true)}
        placeholder="Search for products, brands, and more..."
        autoComplete="off"
      />
      {showSuggestions && suggestions.length > 0 && (
        <Paper
          ref={suggestionsRef}
          elevation={3}
          sx={{
            position: 'absolute',
            top: '100%',
            left: 0,
            right: 0,
            mt: 0.5,
            zIndex: 1200,
          }}
        >
          <List>
            {suggestions.map((suggestion, index) => (
              <ListItem 
                key={index} 
                button 
                onClick={() => handleSuggestionClick(suggestion)}
              >
                <ListItemText primary={suggestion} />
              </ListItem>
            ))}
          </List>
        </Paper>
      )}
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
