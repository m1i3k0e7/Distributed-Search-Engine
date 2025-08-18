
import React from 'react';
import {
  Typography,
  List,
  ListItem,
  Checkbox,
  FormControlLabel,
  Paper
} from '@mui/material';
import { CATEGORIES } from '../constants';

function FilterSidebar({ selectedCategories, onCategoryChange }) {
  const handleToggle = (category) => () => {
    const currentIndex = selectedCategories.indexOf(category);
    const newSelectedCategories = [...selectedCategories];

    if (currentIndex === -1) {
      newSelectedCategories.push(category);
    } else {
      newSelectedCategories.splice(currentIndex, 1);
    }

    onCategoryChange(newSelectedCategories);
  };

  return (
    <Paper 
      elevation={2} 
      sx={{ 
        p: 2, 
        height: '100%'
      }}
    >
      <Typography variant="h6" gutterBottom>
        Categories
      </Typography>
      <List dense>
        {CATEGORIES.map((category) => (
          <ListItem key={category} disablePadding>
            <FormControlLabel
              control={
                <Checkbox
                  checked={selectedCategories.indexOf(category) !== -1}
                  onChange={handleToggle(category)}
                />
              }
              label={category}
              sx={{ width: '100%' }}
            />
          </ListItem>
        ))}
      </List>
    </Paper>
  );
}

export default FilterSidebar;
