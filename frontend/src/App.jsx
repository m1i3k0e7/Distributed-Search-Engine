
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import HomePage from './pages/HomePage';
import ResultsPage from './pages/ResultsPage';

import { ThemeProvider, createTheme, responsiveFontSizes } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import Container from '@mui/material/Container';

let professionalTheme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#000000', // Black
    },
    secondary: {
      main: '#dc004e', // A contrasting pink/red for secondary actions
    },
    background: {
      default: '#F4F6F8', // A very light grey for a soft background
      paper: '#ffffff',   // White for cards and paper elements
    },
  },
  typography: {
    fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
    h1: { fontWeight: 700 },
    h2: { fontWeight: 600 },
    h3: { fontWeight: 600 },
    h4: { fontWeight: 500 },
    h5: { fontWeight: 500 },
    h6: { fontWeight: 500 },
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: '8px',
          textTransform: 'none',
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: '12px',
          transition: 'box-shadow 0.3s ease-in-out, transform 0.3s ease-in-out',
          '&:hover': {
            transform: 'translateY(-4px)',
            boxShadow: '0 4px 20px 0 rgba(0,0,0,0.1)',
          }
        }
      }
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: '12px',
        }
      }
    }
  },
});

professionalTheme = responsiveFontSizes(professionalTheme);

function App() {
  return (
    <ThemeProvider theme={professionalTheme}>
      <CssBaseline />
      <Container maxWidth="xl">
        <Router>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/search" element={<ResultsPage />} />
          </Routes>
        </Router>
      </Container>
    </ThemeProvider>
  );
}

export default App;
