# HTML construction kit

**NOTE**  
The library is still in flux - the api is not yet stable and it is not tested, just written.

HTML construction kit (`hkc`) builds upon [golang.org/x/net/html](golang.org/x/net/html) (`html`) and simpifies tree construction and modification.

While `html` is great for parsing and rendering, its api is rather unfriendly when one wants to build or modify a html document.  
The nodes carry references to their parents, both siblings and the first and last children. It's easy to forget to update one and all referenced nodes potentially have to be updated, too. Modification and querying of attributes is also unwieldly.

`hkc` stores a minimal representation of a node and relies on a `Cursor` to provide context on navigation. Each node only references its children and can easily be moved around.
