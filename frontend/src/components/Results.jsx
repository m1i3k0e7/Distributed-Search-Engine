import React from 'react';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardMedia from '@mui/material/CardMedia';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

function ProductCard({ product }) {
  const imageUrl = product.Image && product.Image.startsWith('http') 
    ? product.Image 
    : `https://via.placeholder.com/400x300.png?text=${encodeURIComponent(product.Name || 'No Image')}`;

  return (
    <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <CardMedia
        component="img"
        height="200"
        image={imageUrl}
        alt={product.Name}
      />
      <CardContent sx={{ flexGrow: 1 }}>
        <Typography gutterBottom variant="h6" component="div" sx={{ fontWeight: 500 }}>
          {product.Name || 'No Title'}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {product.Category || 'Uncategorized'}
        </Typography>
      </CardContent>
      <Box sx={{ p: 2, pt: 0, mt: 'auto' }}>
        <Typography variant="h5" component="div" color="text.primary" fontWeight="bold">
          ${product.DiscountPrice ? product.DiscountPrice.toFixed(2) : 'N/A'}
          {product.ActualPrice && 
            <Typography component="span" sx={{ ml: 1, textDecoration: 'line-through' }} color="text.secondary">
              ${product.ActualPrice.toFixed(2)}
            </Typography>
          }
        </Typography>
      </Box>
    </Card>
  );
}

function Results({ products }) {
  if (!products || products.length === 0) {
    return <Typography align="center" sx={{ mt: 8, fontSize: '1.2rem' }}>No products found. Please try a different search or filter.</Typography>;
  }

  return (
    <Box sx={{
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fill, minmax(250px, 1fr))',
      gap: '24px'
    }}>
      {products.map((product) => (
        <ProductCard key={product.Id} product={product} />
      ))}
    </Box>
  );
}

export default Results;