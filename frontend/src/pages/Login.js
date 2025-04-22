import React, { useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import { loginUser } from '../utils/api';
import { AuthContext } from '../context/AuthContext';
import AuthForm from '../components/AuthForm';
import AuthLayout from '../layouts/AuthLayout'; // ✅ Import layout

const Login = () => {
    const [identifier, setIdentifier] = useState('');
    const [password, setPassword] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');
    const { login } = useContext(AuthContext);
    const navigate = useNavigate();

    const handleLogin = async () => {
        if (!identifier || !password) {
            setError('Please enter both username/email and password');
            return;
        }

        setIsLoading(true);
        setError(''); // Clear any previous errors

        try {
            const response = await loginUser({ identifier, password });
            login(response.data);
            navigate('/dashboard');
        } catch (error) {
            const errorMessage = error.response?.data?.error || 'Invalid credentials';
            setError(errorMessage);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <AuthLayout>
            <AuthForm
                title="Login"
                fields={[
                    {
                        label: "Username/Email",
                        type: "text",
                        value: identifier,
                        onChange: (e) => setIdentifier(e.target.value)
                    },
                    {
                        label: "Password",
                        type: "password",
                        value: password,
                        onChange: (e) => setPassword(e.target.value)
                    },
                ]}
                onSubmit={handleLogin}
                isLoading={isLoading}
                error={error}
                submitButtonText="Login"
                footer={
                    <div>
                        <p><a href="/forgot-password">Forgot Password?</a></p>
                        <p>Don't have an account? <a href="/signup">Signup</a></p>
                    </div>
                }
            />
        </AuthLayout>
    );
};

export default Login;
