import React from "react";
import { connect } from 'react-redux';
import './album.css'

class ProfilePage extends React.Component {

  render (){
    let { profile } = this.props
    console.log(profile)
    return (
        <main role="main">
            <section className="jumbotron text-center hero-bg" style={{"backgroundImage": `url(${profile.BackgroundPhoto})`, "bottom": "0px"}}></section>
                <div className="hero-img" >
                    <img src={profile.ProfilePhoto} width="125px" className="img rounded-circle img-thumbnail" alt="Profile"/>
                    <h1 className="jumbotron-heading">{profile.Photographer}</h1>
                    <ul className="icons">
                    { profile.Twitter && <li><a href={"https://twitter.com/"+profile.Twitter} className="fab fa-twitter">twitter</a></li> }
                    { profile.Instagram && <li><a href={"https://instagram.com/"+profile.Instagram} className="fab fa-instagram">Instagram</a></li> }
                    { profile.Website && <li><a href={profile.Website} className="fab fa-globe">Website</a></li> }
                    </ul>
                </div>
        </main>
    );
    }
}

const mapToProps = (state) => {
    const profile = state.ProfileReducer.profile;
    return {
        profile,
    };
  }
  export default connect(mapToProps)(ProfilePage)
  

