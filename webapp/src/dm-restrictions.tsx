import React, {useState, useEffect} from 'react';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/users';
import {useSelector} from 'react-redux';
import {AlertCircle} from 'lucide-react';

const DMRestrictionsComponent = () => {
    const [canSendDMs, setCanSendDMs] = useState(true);
    const currentUser = useSelector(getCurrentUser);

    useEffect(() => {
        const fetchRestrictions = async () => {
            try {
                const response = await fetch('/plugins/com.mattermost.dm-control-plugin/restrictions');
                const data = await response.json();
                setCanSendDMs(data.canSendDMs);
            } catch (error) {
                // eslint-disable-next-line no-console
                console.error('Failed to fetch DM restrictions:', error);
            }
        };

        fetchRestrictions();
    }, [currentUser.id]);

    if (canSendDMs) {
        return <></>;
    }

    return (
        <div className='fixed bottom-4 right-4 bg-red-600 p-4 rounded-lg shadow-lg text-white flex items-center'>
            <AlertCircle className='mr-2'/>
            <span>{'You are not allowed to send direct messages.'}</span>
        </div>
    );
};

export default DMRestrictionsComponent;
