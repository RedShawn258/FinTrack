import React, { useState } from 'react';
import { resetPassword } from '../utils/api';
import AuthForm from '../components/AuthForm';
import { useNavigate } from 'react-router-dom';

const ForgotPassword = () => {
    const [userData, setUserData] = useState({ identifier: '', newPassword: '', confirmPassword: '' });
    const navigate = useNavigate();

    const handleResetPassword = async () => {
        if (!userData.identifier || !userData.newPassword || !userData.confirmPassword) {
            alert('Failed to reset password');
            return;
        }
        
        if (userData.newPassword !== userData.confirmPassword) {
            alert('Passwords do not match');
            return;
        }

        try {
            await resetPassword(userData);
            alert('Password reset successful');
            navigate('/login');
        } catch (error) {
            alert(error.response?.data?.error || 'Failed to reset password');
        }
    };

    return (
        <AuthForm
            title="Reset Password"
            fields={[
                { label: "Username/Email", type: "text", value: userData.identifier, onChange: (e) => setUserData({ ...userData, identifier: e.target.value }) },
                { label: "New Password", type: "password", value: userData.newPassword, onChange: (e) => setUserData({ ...userData, newPassword: e.target.value }) },
                { label: "Confirm Password", type: "password", value: userData.confirmPassword, onChange: (e) => setUserData({ ...userData, confirmPassword: e.target.value }) },
            ]}
            onSubmit={handleResetPassword}
            footer={<p><a href="/login">Back to Login</a></p>}
        />
    );
};

export default ForgotPassword;
