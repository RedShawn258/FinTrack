import React, { useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import { loginUser } from '../utils/api';
import { AuthContext } from '../context/AuthContext';
import AuthForm from '../components/AuthForm';

const Login = () => {
    const [identifier, setIdentifier] = useState('');
    const [password, setPassword] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const { login } = useContext(AuthContext);
    const navigate = useNavigate();

    const handleLogin = async () => {
        if (!identifier || !password) {
            alert('Please enter both username/email and password');
            return;
        }

        setIsLoading(true);
        try {
            const response = await loginUser({ identifier, password });
            login(response.data);
            
            // This will trigger the data prefetching in AuthContext
            // Then navigate to dashboard
            navigate('/dashboard');
        } catch (error) {
            alert(error.response?.data?.error || 'Login failed');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <AuthForm
            title="Login"
            fields={[
                { label: "Username/Email", type: "text", value: identifier, onChange: (e) => setIdentifier(e.target.value) },
                { label: "Password", type: "password", value: password, onChange: (e) => setPassword(e.target.value) },
            ]}
            onSubmit={handleLogin}
            isLoading={isLoading}
            footer={
                <div>
                    <p><a href="/forgot-password">Forgot Password?</a></p>
                    <p>Don't have an account? <a href="/signup">Signup</a></p>
                </div>
            }
        />
    );
};

export default Login;
