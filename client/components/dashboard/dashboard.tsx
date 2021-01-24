import React from 'react';
import { observer } from 'mobx-react-lite';
import { IApp } from '@/state/models';
import { Service } from './service/service';
import './dashboard.scss';

type TProps = {
    app: IApp;
}

export const Dashboard = observer((props: TProps) => {
    return <div className="dashboard__component">
        <h1>Dashboard</h1>
        <section>
            <h2>Services</h2>
            {props.app.services.map((service, index) =>
                <Service key={index} service={service} />)}
        </section>

        
    </div>;
})

export default Dashboard;