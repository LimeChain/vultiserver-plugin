import { useState } from 'react';
import './App.css'
import Modal from './modules/core/components/ui/modal/Modal';
import DCAPluginPolicyForm from './modules/dca-plugin/components/DCAPluginPolicyForm';

const App = () => {
    const [isModalOpen, setModalOpen] = useState(false);

    return (
        <>
            <DCAPluginPolicyForm />
            {!isModalOpen && <button onClick={() => setModalOpen(true)}>Form in modal</button>}
            <Modal isOpen={isModalOpen} onClose={() => setModalOpen(false)} />
        </>
    );
};

export default App;
