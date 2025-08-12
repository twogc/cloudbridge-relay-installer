// import React from 'react';  // Not needed in React 18 with JSX transform
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import { CssBaseline } from '@mui/material';
import { Provider } from 'react-redux';
import { QueryClient, QueryClientProvider } from 'react-query';
import { SnackbarProvider } from 'notistack';

import { store } from './store';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Events from './pages/Events';
import Alerts from './pages/Alerts';
import CorrelationRules from './pages/CorrelationRules';
import Analytics from './pages/Analytics';
import Settings from './pages/Settings';
import Login from './pages/Login';

// Create theme
const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
    background: {
      default: '#0a0a0a',
      paper: '#1a1a1a',
    },
  },
  typography: {
    fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
  },
  components: {
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: '#1a1a1a',
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          backgroundColor: '#1a1a1a',
        },
      },
    },
  },
});

// Create query client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  return (
    <Provider store={store}>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider theme={theme}>
          <CssBaseline />
          <SnackbarProvider maxSnack={3}>
            <Router>
              <Routes>
                <Route path="/login" element={<Login />} />
                <Route path="/" element={<Layout />}>
                  <Route index element={<Dashboard />} />
                  <Route path="events" element={<Events />} />
                  <Route path="alerts" element={<Alerts />} />
                  <Route path="correlation-rules" element={<CorrelationRules />} />
                  <Route path="analytics" element={<Analytics />} />
                  <Route path="settings" element={<Settings />} />
                </Route>
              </Routes>
            </Router>
          </SnackbarProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </Provider>
  );
}

export default App; 