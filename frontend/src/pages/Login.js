import React, { useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import { loginUser } from '../utils/api';
import { AuthContext } from '../context/AuthContext';
import AuthForm from '../components/AuthForm';

const Login = () => {
    const [identifier, setIdentifier] = useState('');
    const [password, setPassword] = useState('');
    const { login } = useContext(AuthContext);
    const navigate = useNavigate();

    const handleLogin = async () => {
        try {
            const response = await loginUser({ identifier, password });
            login(response.data);
            navigate('/dashboard'); // Redirect after successful login
        } catch (error) {
            alert(error.response?.data?.error || 'Login failed');
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
