import React from "react";
import { connect } from 'react-redux';
import {config} from '../store'
import {Link} from "react-router-dom";
import placeholder from "../img/placeholder.png"
class AlbumsPage extends React.PureComponent {

  render () {
    console.log(this.props.collections)
  return (
    <main>
        <div className="py-5 bg-light">
            <div className="container">
                <div className="row">
                  {this.props.collections.map((value, index) => {
                    return (
                      <div className="col-md-4" key={value.id}>
                            <div className="card mb-4 shadow-sm">
                                <Link to={"album/"+value.id}>
                                {
                                  value.profile_image === ""
                                  ? <img src={placeholder} alt={value.name} width="100%" height="250px" style={{"objectFit": "cover"}}/>
                                  : <img src={config.imageUrl+value.profile_image} alt={value.name} width="100%" height="250px" style={{"objectFit": "cover"}}/>
                                }
                                </Link>
                                <div className="card-body">
                                    <p className="card-text text-center">{value.name}</p>
                                </div>
                            </div>
                        </div>
                    )
                  })}
                </div>
            </div>
        </div>
    </main>
  );
                }
}
const mapToProps = (state) =>{
  //const collections = state.CollectionsReducer.collections;
  var collections = Object.keys(state.CollectionsReducer.collections).map(function(key) {
    return state.CollectionsReducer.collections[key]
  });
  console.log(collections)



  return {
    collections,
  };
}
export default connect(mapToProps)(AlbumsPage)

