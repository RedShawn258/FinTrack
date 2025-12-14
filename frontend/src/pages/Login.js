import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { loginUser } from '../store/slices/authSlice';
import AuthForm from '../components/AuthForm';
import AuthLayout from '../layouts/AuthLayout';

const Login = () => {
    const [identifier, setIdentifier] = useState('');
    const [password, setPassword] = useState('');
    const dispatch = useDispatch();
    const navigate = useNavigate();
    const { isLoading, error, isAuthenticated } = useSelector((state) => state.auth);

    // Redirect if already authenticated
    useEffect(() => {
        if (isAuthenticated) {
            navigate('/dashboard');
        }
    }, [isAuthenticated, navigate]);

    const handleLogin = async () => {
        if (!identifier || !password) {
            return;
        }

        try {
            const result = await dispatch(loginUser({ identifier, password })).unwrap();
            if (result) {
                navigate('/dashboard');
            }
        } catch (error) {
            // Error is handled by Redux state
            console.error('Login failed:', error);
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
