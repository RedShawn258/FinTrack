import React, { useState } from 'react';
import { registerUser } from '../utils/api';
import AuthForm from '../components/AuthForm';
import { useNavigate } from 'react-router-dom';

const Signup = () => {
    const [userData, setUserData] = useState({ username: '', email: '', password: '', confirmPassword: '' });
    const navigate = useNavigate();

    const handleSignup = async () => {
        if (userData.password !== userData.confirmPassword) {
            alert('Passwords do not match');
            return;
        }
        try {
            await registerUser(userData);
            alert('Signup successful');
            navigate('/login');
        } catch (error) {
            alert(error.response?.data?.error || 'Signup failed');
        }
    };

    return (
        <AuthForm
            title="Signup"
            fields={[
                { label: "Username", type: "text", value: userData.username, onChange: (e) => setUserData({ ...userData, username: e.target.value }) },
                { label: "Email", type: "email", value: userData.email, onChange: (e) => setUserData({ ...userData, email: e.target.value }) },
                { label: "Password", type: "password", value: userData.password, onChange: (e) => setUserData({ ...userData, password: e.target.value }) },
                { label: "Confirm Password", type: "password", value: userData.confirmPassword, onChange: (e) => setUserData({ ...userData, confirmPassword: e.target.value }) },
            ]}
            onSubmit={handleSignup}
            footer={<p>Already have an account? <a href="/login">Login</a></p>}
        />
    );
};

export default Signup;
