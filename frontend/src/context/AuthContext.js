import { createContext, useState, useEffect, useCallback } from 'react';
import { getBudgets, getCategories, getTransactions } from '../utils/api';

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [dashboardData, setDashboardData] = useState({
        budgets: [],
        categories: [],
        transactions: []
    });
    const [isInitialized, setIsInitialized] = useState(false);

    const prefetchDashboardData = async (token) => {
        if (!token) return;
        
        try {
            const [budRes, catRes, txRes] = await Promise.all([
                getBudgets(token),
                getCategories(token),
                getTransactions(token)
            ]);
            
            setDashboardData({
                budgets: budRes.data.budgets,
                categories: catRes.data.categories,
                transactions: txRes.data.transactions
            });
            
            // Store in sessionStorage for persistence
            sessionStorage.setItem('dashboardData', JSON.stringify({
                budgets: budRes.data.budgets,
                categories: catRes.data.categories,
                transactions: txRes.data.transactions,
                timestamp: Date.now()
            }));
        } catch (error) {
            console.error('Failed to prefetch dashboard data', error);
        }
    };

    useEffect(() => {
        // Load user data from localStorage
        const storedUser = localStorage.getItem('user');
        
        if (storedUser) {
            try {
                const userData = JSON.parse(storedUser);
                setUser(userData);
                
                // Try to load cached dashboard data first
                const cachedData = sessionStorage.getItem('dashboardData');
                if (cachedData) {
                    try {
                        const parsedData = JSON.parse(cachedData);
                        setDashboardData({
                            budgets: parsedData.budgets || [],
                            categories: parsedData.categories || [],
                            transactions: parsedData.transactions || []
                        });
                    } catch (e) {
                        console.error('Error parsing cached dashboard data', e);
                        // Continue with empty data if parsing fails
                    }
                }
                
                // Then refresh data in the background if we have a token
                if (userData.token) {
                    prefetchDashboardData(userData.token);
                }
            } catch (e) {
                console.error('Error loading user data from localStorage', e);
                // Clear corrupted data
                localStorage.removeItem('user');
            }
        }
        
        setIsInitialized(true);
    }, []);

    const login = (userData) => {
        // Ensure userData has all required fields
        if (!userData || !userData.token) {
            console.error('Invalid user data for login', userData);
            return;
        }
        
        setUser(userData);
        // Store in localStorage for persistence across refreshes
        localStorage.setItem('user', JSON.stringify(userData));
        
        // Prefetch dashboard data immediately upon login
        prefetchDashboardData(userData.token);
    };

    const logout = () => {
        setUser(null);
        setDashboardData({
            budgets: [],
            categories: [],
            transactions: []
        });
        // Remove both storages
        localStorage.removeItem('user');
        sessionStorage.removeItem('dashboardData');
    };

    // Function to update dashboard data after changes - memoize to avoid infinite re-renders
    const refreshDashboardData = useCallback(async () => {
        if (user && user.token) {
            await prefetchDashboardData(user.token);
        }
    }, [user]);

    return (
        <AuthContext.Provider value={{ 
            user, 
            login, 
            logout, 
            dashboardData, 
            refreshDashboardData,
            isInitialized 
        }}>
            {children}
        </AuthContext.Provider>
    );
};
